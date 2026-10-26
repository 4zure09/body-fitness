package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/4zure09/body-fitness/internal/modules/auth/domain"
	"github.com/4zure09/body-fitness/internal/modules/auth/port"
	"github.com/google/uuid"
)

type Service struct {
	repository        port.Repository
	passwords         port.PasswordHasher
	tokens            port.TokenManager
	dummyPasswordHash string
	now               func() time.Time
	newID             func() string
}

type RegisterInput struct {
	Email       string
	Password    string
	DisplayName string
}

type LoginInput struct {
	Email    string
	Password string
}

type AuthResult struct {
	User   UserView  `json:"user"`
	Tokens TokenPair `json:"tokens"`
}

type UserView struct {
	ID          string        `json:"id"`
	Email       string        `json:"email"`
	DisplayName string        `json:"display_name"`
	Roles       []domain.Role `json:"roles"`
}

type TokenPair struct {
	AccessToken      string    `json:"access_token"`
	RefreshToken     string    `json:"refresh_token"`
	TokenType        string    `json:"token_type"`
	AccessExpiresAt  time.Time `json:"access_expires_at"`
	RefreshExpiresAt time.Time `json:"refresh_expires_at"`
}

func NewService(
	repository port.Repository,
	passwords port.PasswordHasher,
	tokens port.TokenManager,
) (*Service, error) {
	dummyHash, err := passwords.Hash("dummy-password-used-only-for-timing")
	if err != nil {
		return nil, fmt.Errorf("create dummy password hash: %w", err)
	}
	return &Service{
		repository:        repository,
		passwords:         passwords,
		tokens:            tokens,
		dummyPasswordHash: dummyHash,
		now:               time.Now,
		newID:             uuid.NewString,
	}, nil
}

func (s *Service) Register(ctx context.Context, input RegisterInput) (AuthResult, error) {
	user, err := s.newRegisteredUser(input)
	if err != nil {
		return AuthResult{}, err
	}

	issued, err := s.tokens.IssuePair(user)
	if err != nil {
		return AuthResult{}, fmt.Errorf("issue token pair: %w", err)
	}
	if err := s.repository.CreateUser(ctx, user, s.newRefreshSession(user.ID, issued)); err != nil {
		return AuthResult{}, err
	}
	return newAuthResult(user, issued), nil
}

func (s *Service) Login(ctx context.Context, input LoginInput) (AuthResult, error) {
	email, err := domain.NormalizeEmail(input.Email)
	if err != nil || domain.ValidatePassword(input.Password) != nil {
		return AuthResult{}, domain.ErrInvalidCredentials
	}

	user, err := s.repository.FindUserByEmail(ctx, email)
	if errors.Is(err, domain.ErrNotFound) {
		_, _ = s.passwords.Verify(s.dummyPasswordHash, input.Password)
		return AuthResult{}, domain.ErrInvalidCredentials
	}
	if err != nil {
		return AuthResult{}, err
	}
	if !user.IsActive() {
		return AuthResult{}, domain.ErrAccountDisabled
	}

	valid, err := s.passwords.Verify(user.PasswordHash, input.Password)
	if err != nil {
		return AuthResult{}, fmt.Errorf("verify password: %w", err)
	}
	if !valid {
		return AuthResult{}, domain.ErrInvalidCredentials
	}

	issued, err := s.tokens.IssuePair(user)
	if err != nil {
		return AuthResult{}, fmt.Errorf("issue token pair: %w", err)
	}
	if err := s.repository.StoreRefreshToken(ctx, s.newRefreshSession(user.ID, issued)); err != nil {
		return AuthResult{}, err
	}
	return newAuthResult(user, issued), nil
}

func (s *Service) Refresh(ctx context.Context, rawRefreshToken string) (AuthResult, error) {
	claims, err := s.tokens.ParseRefresh(rawRefreshToken)
	if err != nil {
		return AuthResult{}, domain.ErrInvalidToken
	}

	currentHash := s.tokens.Fingerprint(rawRefreshToken)
	user, err := s.repository.FindUserByActiveRefreshToken(ctx, currentHash, s.now().UTC())
	if errors.Is(err, domain.ErrNotFound) {
		return AuthResult{}, domain.ErrInvalidToken
	}
	if err != nil {
		return AuthResult{}, err
	}
	if user.ID != claims.UserID || !user.IsActive() {
		return AuthResult{}, domain.ErrInvalidToken
	}

	issued, err := s.tokens.IssuePair(user)
	if err != nil {
		return AuthResult{}, fmt.Errorf("issue token pair: %w", err)
	}
	now := s.now().UTC()
	if err := s.repository.RotateRefreshToken(
		ctx,
		currentHash,
		s.newRefreshSession(user.ID, issued),
		now,
	); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return AuthResult{}, domain.ErrInvalidToken
		}
		return AuthResult{}, err
	}
	return newAuthResult(user, issued), nil
}

func (s *Service) Logout(ctx context.Context, rawRefreshToken string) error {
	if _, err := s.tokens.ParseRefresh(rawRefreshToken); err != nil {
		return domain.ErrInvalidToken
	}
	return s.repository.RevokeRefreshToken(
		ctx,
		s.tokens.Fingerprint(rawRefreshToken),
		s.now().UTC(),
	)
}

func (s *Service) CurrentUser(ctx context.Context, userID string) (UserView, error) {
	user, err := s.repository.FindUserByID(ctx, userID)
	if err != nil {
		return UserView{}, err
	}
	if !user.IsActive() {
		return UserView{}, domain.ErrAccountDisabled
	}
	return newUserView(user), nil
}

func (s *Service) newRegisteredUser(input RegisterInput) (domain.User, error) {
	email, err := domain.NormalizeEmail(input.Email)
	if err != nil {
		return domain.User{}, err
	}
	if err := domain.ValidatePassword(input.Password); err != nil {
		return domain.User{}, err
	}
	displayName, err := domain.NormalizeDisplayName(input.DisplayName)
	if err != nil {
		return domain.User{}, err
	}
	passwordHash, err := s.passwords.Hash(input.Password)
	if err != nil {
		return domain.User{}, fmt.Errorf("hash password: %w", err)
	}
	return domain.User{
		ID:           s.newID(),
		Email:        email,
		PasswordHash: passwordHash,
		DisplayName:  displayName,
		Status:       domain.UserStatusActive,
		Roles:        []domain.Role{domain.RoleUser},
		CreatedAt:    s.now().UTC(),
	}, nil
}

func (s *Service) newRefreshSession(userID string, issued port.IssuedTokenPair) port.RefreshSession {
	return port.RefreshSession{
		ID:        issued.RefreshTokenID,
		UserID:    userID,
		TokenHash: s.tokens.Fingerprint(issued.RefreshToken),
		ExpiresAt: issued.RefreshExpiresAt,
	}
}

func newAuthResult(user domain.User, issued port.IssuedTokenPair) AuthResult {
	return AuthResult{
		User: newUserView(user),
		Tokens: TokenPair{
			AccessToken:      issued.AccessToken,
			RefreshToken:     issued.RefreshToken,
			TokenType:        "Bearer",
			AccessExpiresAt:  issued.AccessExpiresAt,
			RefreshExpiresAt: issued.RefreshExpiresAt,
		},
	}
}

func newUserView(user domain.User) UserView {
	return UserView{
		ID:          user.ID,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		Roles:       append([]domain.Role(nil), user.Roles...),
	}
}
