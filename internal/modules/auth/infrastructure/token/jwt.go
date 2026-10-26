package token

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/4zure09/body-fitness/internal/modules/auth/domain"
	"github.com/4zure09/body-fitness/internal/modules/auth/port"
	"github.com/4zure09/body-fitness/internal/platform/config"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	tokenUseAccess  = "access"
	tokenUseRefresh = "refresh"
)

type Manager struct {
	secret     []byte
	issuer     string
	audience   string
	accessTTL  time.Duration
	refreshTTL time.Duration
	now        func() time.Time
	newID      func() string
}

type claims struct {
	Roles    []domain.Role `json:"roles,omitempty"`
	TokenUse string        `json:"token_use"`
	jwt.RegisteredClaims
}

func NewManager(cfg config.Auth) Manager {
	return Manager{
		secret:     []byte(cfg.JWTSecret),
		issuer:     cfg.Issuer,
		audience:   cfg.Audience,
		accessTTL:  cfg.AccessTTL,
		refreshTTL: cfg.RefreshTTL,
		now:        time.Now,
		newID:      uuid.NewString,
	}
}

func (m Manager) IssuePair(user domain.User) (port.IssuedTokenPair, error) {
	now := m.now().UTC()
	accessID := m.newID()
	refreshID := m.newID()
	accessExpiry := now.Add(m.accessTTL)
	refreshExpiry := now.Add(m.refreshTTL)

	accessToken, err := m.sign(user.ID, accessID, user.Roles, tokenUseAccess, now, accessExpiry)
	if err != nil {
		return port.IssuedTokenPair{}, fmt.Errorf("sign access token: %w", err)
	}
	refreshToken, err := m.sign(user.ID, refreshID, nil, tokenUseRefresh, now, refreshExpiry)
	if err != nil {
		return port.IssuedTokenPair{}, fmt.Errorf("sign refresh token: %w", err)
	}
	return port.IssuedTokenPair{
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		RefreshTokenID:   refreshID,
		AccessExpiresAt:  accessExpiry,
		RefreshExpiresAt: refreshExpiry,
	}, nil
}

func (m Manager) ParseAccess(rawToken string) (port.Claims, error) {
	return m.parse(rawToken, tokenUseAccess)
}

func (m Manager) ParseRefresh(rawToken string) (port.Claims, error) {
	return m.parse(rawToken, tokenUseRefresh)
}

func (m Manager) Fingerprint(rawToken string) string {
	sum := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(sum[:])
}

func (m Manager) sign(
	userID string,
	tokenID string,
	roles []domain.Role,
	tokenUse string,
	issuedAt time.Time,
	expiresAt time.Time,
) (string, error) {
	tokenClaims := claims{
		Roles:    append([]domain.Role(nil), roles...),
		TokenUse: tokenUse,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   userID,
			Audience:  jwt.ClaimStrings{m.audience},
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(issuedAt),
			NotBefore: jwt.NewNumericDate(issuedAt),
			ID:        tokenID,
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, tokenClaims).SignedString(m.secret)
}

func (m Manager) parse(rawToken, expectedUse string) (port.Claims, error) {
	tokenClaims := &claims{}
	token, err := jwt.ParseWithClaims(
		rawToken,
		tokenClaims,
		func(token *jwt.Token) (any, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, domain.ErrInvalidToken
			}
			return m.secret, nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(m.issuer),
		jwt.WithAudience(m.audience),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
		jwt.WithTimeFunc(m.now),
	)
	if err != nil || !token.Valid || tokenClaims.TokenUse != expectedUse {
		return port.Claims{}, domain.ErrInvalidToken
	}
	if tokenClaims.Subject == "" || tokenClaims.ID == "" || tokenClaims.ExpiresAt == nil {
		return port.Claims{}, domain.ErrInvalidToken
	}
	return port.Claims{
		UserID:    tokenClaims.Subject,
		TokenID:   tokenClaims.ID,
		Roles:     append([]domain.Role(nil), tokenClaims.Roles...),
		ExpiresAt: tokenClaims.ExpiresAt.Time,
	}, nil
}
