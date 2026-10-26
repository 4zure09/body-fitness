package port

import (
	"time"

	"github.com/4zure09/body-fitness/internal/modules/auth/domain"
)

type Claims struct {
	UserID    string
	TokenID   string
	Roles     []domain.Role
	ExpiresAt time.Time
}

type IssuedTokenPair struct {
	AccessToken      string
	RefreshToken     string
	RefreshTokenID   string
	AccessExpiresAt  time.Time
	RefreshExpiresAt time.Time
}

type TokenManager interface {
	IssuePair(user domain.User) (IssuedTokenPair, error)
	ParseAccess(rawToken string) (Claims, error)
	ParseRefresh(rawToken string) (Claims, error)
	Fingerprint(rawToken string) string
}
