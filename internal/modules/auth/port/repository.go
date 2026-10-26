package port

import (
	"context"
	"time"

	"github.com/4zure09/body-fitness/internal/modules/auth/domain"
)

type RefreshSession struct {
	ID            string
	UserID        string
	TokenHash     string
	ExpiresAt     time.Time
	RevokedAt     *time.Time
	ReplacementID *string
}

type Repository interface {
	CreateUser(ctx context.Context, user domain.User, session RefreshSession) error
	FindUserByEmail(ctx context.Context, email string) (domain.User, error)
	FindUserByID(ctx context.Context, userID string) (domain.User, error)
	FindUserByActiveRefreshToken(ctx context.Context, tokenHash string, now time.Time) (domain.User, error)
	GrantRoleByEmail(ctx context.Context, email string, role domain.Role, grantedAt time.Time) error
	StoreRefreshToken(ctx context.Context, session RefreshSession) error
	RotateRefreshToken(ctx context.Context, currentHash string, replacement RefreshSession, revokedAt time.Time) error
	RevokeRefreshToken(ctx context.Context, tokenHash string, revokedAt time.Time) error
}
