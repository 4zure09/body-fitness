package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/4zure09/body-fitness/internal/modules/auth/domain"
	"github.com/4zure09/body-fitness/internal/modules/auth/port"
)

type repositoryStub struct {
	createdUser        domain.User
	createdSession     port.RefreshSession
	userByEmail        domain.User
	userByEmailErr     error
	userByID           domain.User
	userByIDErr        error
	refreshUser        domain.User
	refreshUserErr     error
	storedSession      port.RefreshSession
	rotatedCurrentHash string
	rotatedSession     port.RefreshSession
	revokedHash        string
	returnErr          error
}

func (s *repositoryStub) CreateUser(_ context.Context, user domain.User, session port.RefreshSession) error {
	s.createdUser = user
	s.createdSession = session
	return s.returnErr
}

func TestGrantRoleRejectsUnsupportedRole(t *testing.T) {
	repository := &repositoryStub{}
	service := newTestService(t, repository, &tokenManagerStub{})

	err := service.GrantRole(context.Background(), "user@example.com", domain.Role("owner"))
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("GrantRole() error = %v, want invalid input", err)
	}
}

func (s *repositoryStub) FindUserByEmail(context.Context, string) (domain.User, error) {
	return s.userByEmail, s.userByEmailErr
}

func (s *repositoryStub) FindUserByID(context.Context, string) (domain.User, error) {
	return s.userByID, s.userByIDErr
}

func (s *repositoryStub) FindUserByActiveRefreshToken(context.Context, string, time.Time) (domain.User, error) {
	return s.refreshUser, s.refreshUserErr
}

func (s *repositoryStub) StoreRefreshToken(_ context.Context, session port.RefreshSession) error {
	s.storedSession = session
	return s.returnErr
}

func (s *repositoryStub) GrantRoleByEmail(context.Context, string, domain.Role, time.Time) error {
	return s.returnErr
}

func (s *repositoryStub) RotateRefreshToken(
	_ context.Context,
	currentHash string,
	replacement port.RefreshSession,
	_ time.Time,
) error {
	s.rotatedCurrentHash = currentHash
	s.rotatedSession = replacement
	return s.returnErr
}

func (s *repositoryStub) RevokeRefreshToken(_ context.Context, tokenHash string, _ time.Time) error {
	s.revokedHash = tokenHash
	return s.returnErr
}

type passwordStub struct{}

func (passwordStub) Hash(password string) (string, error) {
	return "hashed:" + password, nil
}

func (passwordStub) Verify(hash, password string) (bool, error) {
	return hash == "hashed:"+password, nil
}

type tokenManagerStub struct {
	issuedFor     domain.User
	refreshClaims port.Claims
	parseError    error
	sequence      int
}

func (s *tokenManagerStub) IssuePair(user domain.User) (port.IssuedTokenPair, error) {
	s.issuedFor = user
	s.sequence++
	return port.IssuedTokenPair{
		AccessToken:      "access-token",
		RefreshToken:     "refresh-token-" + string(rune('0'+s.sequence)),
		RefreshTokenID:   "refresh-id-" + string(rune('0'+s.sequence)),
		AccessExpiresAt:  time.Date(2026, 10, 26, 10, 15, 0, 0, time.UTC),
		RefreshExpiresAt: time.Date(2026, 11, 25, 10, 0, 0, 0, time.UTC),
	}, nil
}

func (s *tokenManagerStub) ParseAccess(string) (port.Claims, error) {
	return port.Claims{}, errors.New("not used")
}

func (s *tokenManagerStub) ParseRefresh(string) (port.Claims, error) {
	return s.refreshClaims, s.parseError
}

func (s *tokenManagerStub) Fingerprint(rawToken string) string {
	return "fingerprint:" + rawToken
}

func newTestService(t *testing.T, repository *repositoryStub, tokens *tokenManagerStub) *Service {
	t.Helper()
	service, err := NewService(repository, passwordStub{}, tokens)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
	service.newID = func() string { return "user-id" }
	service.now = func() time.Time { return time.Date(2026, 10, 26, 10, 0, 0, 0, time.UTC) }
	return service
}

func TestRegisterCreatesUserWithDefaultRoleAndRefreshSession(t *testing.T) {
	repository := &repositoryStub{}
	tokens := &tokenManagerStub{}
	service := newTestService(t, repository, tokens)

	result, err := service.Register(context.Background(), RegisterInput{
		Email:       " User@Example.com ",
		Password:    "password123",
		DisplayName: " Fitness User ",
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if repository.createdUser.Email != "user@example.com" {
		t.Fatalf("created email = %q", repository.createdUser.Email)
	}
	if len(repository.createdUser.Roles) != 1 || repository.createdUser.Roles[0] != domain.RoleUser {
		t.Fatalf("created roles = %+v", repository.createdUser.Roles)
	}
	if repository.createdSession.TokenHash != "fingerprint:refresh-token-1" {
		t.Fatalf("unexpected refresh session: %+v", repository.createdSession)
	}
	if result.Tokens.TokenType != "Bearer" || result.User.ID != "user-id" {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestRegisterRejectsInvalidInputBeforePersistence(t *testing.T) {
	repository := &repositoryStub{}
	service := newTestService(t, repository, &tokenManagerStub{})

	_, err := service.Register(context.Background(), RegisterInput{
		Email:       "invalid",
		Password:    "short",
		DisplayName: "A",
	})
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("Register() error = %v, want invalid input", err)
	}
	if repository.createdUser.ID != "" {
		t.Fatal("invalid registration must not be persisted")
	}
}

func TestLoginReturnsSameErrorForUnknownEmailAndWrongPassword(t *testing.T) {
	unknownRepository := &repositoryStub{userByEmailErr: domain.ErrNotFound}
	unknownService := newTestService(t, unknownRepository, &tokenManagerStub{})
	_, unknownErr := unknownService.Login(context.Background(), LoginInput{
		Email: "unknown@example.com", Password: "password123",
	})

	wrongRepository := &repositoryStub{userByEmail: activeUser()}
	wrongService := newTestService(t, wrongRepository, &tokenManagerStub{})
	_, wrongErr := wrongService.Login(context.Background(), LoginInput{
		Email: "user@example.com", Password: "wrong-password",
	})

	if !errors.Is(unknownErr, domain.ErrInvalidCredentials) || !errors.Is(wrongErr, domain.ErrInvalidCredentials) {
		t.Fatalf("unexpected errors: unknown=%v wrong=%v", unknownErr, wrongErr)
	}
}

func TestLoginStoresRefreshSession(t *testing.T) {
	repository := &repositoryStub{userByEmail: activeUser()}
	tokens := &tokenManagerStub{}
	service := newTestService(t, repository, tokens)

	_, err := service.Login(context.Background(), LoginInput{
		Email: "user@example.com", Password: "password123",
	})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if repository.storedSession.UserID != "user-id" || repository.storedSession.TokenHash == "" {
		t.Fatalf("unexpected stored session: %+v", repository.storedSession)
	}
}

func TestRefreshRotatesToken(t *testing.T) {
	repository := &repositoryStub{refreshUser: activeUser()}
	tokens := &tokenManagerStub{refreshClaims: port.Claims{UserID: "user-id", TokenID: "old-id"}}
	service := newTestService(t, repository, tokens)

	_, err := service.Refresh(context.Background(), "old-refresh-token")
	if err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}
	if repository.rotatedCurrentHash != "fingerprint:old-refresh-token" {
		t.Fatalf("rotated hash = %q", repository.rotatedCurrentHash)
	}
	if repository.rotatedSession.ID == "" || repository.rotatedSession.UserID != "user-id" {
		t.Fatalf("unexpected replacement: %+v", repository.rotatedSession)
	}
}

func TestRefreshRejectsTokenForDifferentUser(t *testing.T) {
	repository := &repositoryStub{refreshUser: activeUser()}
	tokens := &tokenManagerStub{refreshClaims: port.Claims{UserID: "another-user"}}
	service := newTestService(t, repository, tokens)

	_, err := service.Refresh(context.Background(), "refresh-token")
	if !errors.Is(err, domain.ErrInvalidToken) {
		t.Fatalf("Refresh() error = %v, want invalid token", err)
	}
}

func TestLogoutRevokesRefreshTokenFingerprint(t *testing.T) {
	repository := &repositoryStub{}
	tokens := &tokenManagerStub{refreshClaims: port.Claims{UserID: "user-id"}}
	service := newTestService(t, repository, tokens)

	if err := service.Logout(context.Background(), "refresh-token"); err != nil {
		t.Fatalf("Logout() error = %v", err)
	}
	if repository.revokedHash != "fingerprint:refresh-token" {
		t.Fatalf("revoked hash = %q", repository.revokedHash)
	}
}

func activeUser() domain.User {
	return domain.User{
		ID:           "user-id",
		Email:        "user@example.com",
		PasswordHash: "hashed:password123",
		DisplayName:  "User",
		Status:       domain.UserStatusActive,
		Roles:        []domain.Role{domain.RoleUser},
	}
}
