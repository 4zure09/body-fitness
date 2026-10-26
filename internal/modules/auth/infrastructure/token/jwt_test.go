package token

import (
	"testing"
	"time"

	"github.com/4zure09/body-fitness/internal/modules/auth/domain"
	"github.com/4zure09/body-fitness/internal/platform/config"
)

func TestManagerIssuesAndParsesTokenPair(t *testing.T) {
	manager := NewManager(config.Auth{
		JWTSecret:  "test-secret-with-at-least-thirty-two-characters",
		Issuer:     "test-issuer",
		Audience:   "test-audience",
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 24 * time.Hour,
	})
	fixedNow := time.Date(2026, 10, 26, 10, 0, 0, 0, time.UTC)
	ids := []string{"access-id", "refresh-id"}
	manager.now = func() time.Time { return fixedNow }
	manager.newID = func() string {
		id := ids[0]
		ids = ids[1:]
		return id
	}

	issued, err := manager.IssuePair(domain.User{
		ID:    "user-id",
		Roles: []domain.Role{domain.RoleUser, domain.RoleAdmin},
	})
	if err != nil {
		t.Fatalf("IssuePair() error = %v", err)
	}
	accessClaims, err := manager.ParseAccess(issued.AccessToken)
	if err != nil {
		t.Fatalf("ParseAccess() error = %v", err)
	}
	if accessClaims.UserID != "user-id" || accessClaims.TokenID != "access-id" {
		t.Fatalf("unexpected access claims: %+v", accessClaims)
	}
	if len(accessClaims.Roles) != 2 || accessClaims.Roles[1] != domain.RoleAdmin {
		t.Fatalf("unexpected access roles: %+v", accessClaims.Roles)
	}
	refreshClaims, err := manager.ParseRefresh(issued.RefreshToken)
	if err != nil {
		t.Fatalf("ParseRefresh() error = %v", err)
	}
	if refreshClaims.TokenID != "refresh-id" || issued.RefreshTokenID != "refresh-id" {
		t.Fatalf("unexpected refresh token id: %+v", refreshClaims)
	}
}

func TestManagerRejectsWrongTokenUse(t *testing.T) {
	manager := NewManager(config.Auth{
		JWTSecret:  "test-secret-with-at-least-thirty-two-characters",
		Issuer:     "test-issuer",
		Audience:   "test-audience",
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 24 * time.Hour,
	})
	issued, err := manager.IssuePair(domain.User{ID: "user-id", Roles: []domain.Role{domain.RoleUser}})
	if err != nil {
		t.Fatalf("IssuePair() error = %v", err)
	}
	if _, err := manager.ParseAccess(issued.RefreshToken); err == nil {
		t.Fatal("expected refresh token to be rejected as access token")
	}
}

func TestFingerprintIsStableAndDoesNotExposeToken(t *testing.T) {
	manager := Manager{}
	first := manager.Fingerprint("secret-token")
	second := manager.Fingerprint("secret-token")
	if first != second {
		t.Fatal("expected stable token fingerprint")
	}
	if first == "secret-token" || len(first) != 64 {
		t.Fatalf("unexpected fingerprint %q", first)
	}
}
