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
