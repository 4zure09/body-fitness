package config

import "testing"

func TestLoadRejectsInvalidConnectionLimits(t *testing.T) {
	t.Setenv("DATABASE_MAX_OPEN_CONNS", "5")
	t.Setenv("DATABASE_MAX_IDLE_CONNS", "6")

	_, err := Load()
	if err == nil {
		t.Fatal("expected invalid connection limits to return an error")
	}
}

func TestLoadUsesHTTPAddressFromEnvironment(t *testing.T) {
	t.Setenv("HTTP_ADDRESS", ":9090")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.HTTP.Address != ":9090" {
		t.Fatalf("HTTP address = %q, want %q", cfg.HTTP.Address, ":9090")
	}
}

func TestLoadUsesRailwayDatabaseURLWhenDSNIsMissing(t *testing.T) {
	t.Setenv("DATABASE_DSN", "")
	t.Setenv("DATABASE_URL", "postgresql://railway-user:secret@postgres.railway.internal:5432/railway")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Database.DSN != "postgresql://railway-user:secret@postgres.railway.internal:5432/railway" {
		t.Fatalf("database DSN = %q, want DATABASE_URL", cfg.Database.DSN)
	}
}

func TestLoadPrefersDatabaseDSNOverDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_DSN", "postgresql://local-override")
	t.Setenv("DATABASE_URL", "postgresql://railway")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Database.DSN != "postgresql://local-override" {
		t.Fatalf("database DSN = %q, want DATABASE_DSN", cfg.Database.DSN)
	}
}

func TestLoadRejectsDefaultJWTSecretOutsideLocal(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("JWT_SECRET", "local-development-secret-change-before-production-2026")

	_, err := Load()
	if err == nil {
		t.Fatal("expected production environment to reject local JWT secret")
	}
}

func TestLoadRejectsAccessTTLNotShorterThanRefreshTTL(t *testing.T) {
	t.Setenv("JWT_ACCESS_TTL", "24h")
	t.Setenv("JWT_REFRESH_TTL", "1h")

	_, err := Load()
	if err == nil {
		t.Fatal("expected invalid JWT TTLs to return an error")
	}
}
