package repository

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/4zure09/body-fitness/internal/modules/auth/domain"
	"github.com/4zure09/body-fitness/internal/modules/auth/port"
	"github.com/4zure09/body-fitness/internal/platform/config"
	"github.com/4zure09/body-fitness/internal/platform/database"
	"github.com/google/uuid"
)

func TestGORMRepositoryAuthLifecycle(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("TEST_DATABASE_DSN is not configured")
	}

	db, err := database.Open(context.Background(), config.Database{
		DSN:             dsn,
		MaxOpenConns:    5,
		MaxIdleConns:    2,
		ConnMaxLifetime: time.Minute,
	})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	t.Cleanup(func() { _ = database.Close(db) })

	repository := NewGORMRepository(db)
	now := time.Now().UTC().Truncate(time.Millisecond)
	userID := uuid.NewString()
	email := "repository-" + userID + "@example.com"
	user := domain.User{
		ID:           userID,
		Email:        email,
		PasswordHash: "encoded-password-hash",
		DisplayName:  "Repository Test",
		Status:       domain.UserStatusActive,
		Roles:        []domain.Role{domain.RoleUser},
		CreatedAt:    now,
	}
	firstSession := port.RefreshSession{
		ID:        uuid.NewString(),
		UserID:    userID,
		TokenHash: "first-" + uuid.NewString(),
		ExpiresAt: now.Add(time.Hour),
	}
	t.Cleanup(func() { _ = db.Exec("DELETE FROM users WHERE id = ?", userID).Error })

	if err := repository.CreateUser(context.Background(), user, firstSession); err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}
	stored, err := repository.FindUserByEmail(context.Background(), email)
	if err != nil {
		t.Fatalf("FindUserByEmail() error = %v", err)
	}
	if stored.ID != userID || len(stored.Roles) != 1 || stored.Roles[0] != domain.RoleUser {
		t.Fatalf("unexpected stored user: %+v", stored)
	}

	duplicate := user
	duplicate.ID = uuid.NewString()
	duplicateSession := firstSession
	duplicateSession.ID = uuid.NewString()
	duplicateSession.UserID = duplicate.ID
	duplicateSession.TokenHash = "duplicate-" + uuid.NewString()
	if err := repository.CreateUser(context.Background(), duplicate, duplicateSession); !errors.Is(err, domain.ErrEmailAlreadyExists) {
		t.Fatalf("duplicate CreateUser() error = %v, want email already exists", err)
	}

	if err := repository.GrantRoleByEmail(context.Background(), email, domain.RoleAdmin, now); err != nil {
		t.Fatalf("GrantRoleByEmail() error = %v", err)
	}
	stored, err = repository.FindUserByID(context.Background(), userID)
	if err != nil {
		t.Fatalf("FindUserByID() error = %v", err)
	}
	if len(stored.Roles) != 2 || stored.Roles[0] != domain.RoleAdmin || stored.Roles[1] != domain.RoleUser {
		t.Fatalf("unexpected roles after grant: %+v", stored.Roles)
	}

	if _, err := repository.FindUserByActiveRefreshToken(context.Background(), firstSession.TokenHash, now); err != nil {
		t.Fatalf("FindUserByActiveRefreshToken() error = %v", err)
	}
	replacement := port.RefreshSession{
		ID:        uuid.NewString(),
		UserID:    userID,
		TokenHash: "replacement-" + uuid.NewString(),
		ExpiresAt: now.Add(2 * time.Hour),
	}
	if err := repository.RotateRefreshToken(context.Background(), firstSession.TokenHash, replacement, now); err != nil {
		t.Fatalf("RotateRefreshToken() error = %v", err)
	}
	if _, err := repository.FindUserByActiveRefreshToken(context.Background(), firstSession.TokenHash, now); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("old refresh token error = %v, want not found", err)
	}
	if err := repository.RevokeRefreshToken(context.Background(), replacement.TokenHash, now); err != nil {
		t.Fatalf("RevokeRefreshToken() error = %v", err)
	}
	if _, err := repository.FindUserByActiveRefreshToken(context.Background(), replacement.TokenHash, now); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("revoked refresh token error = %v, want not found", err)
	}
}
