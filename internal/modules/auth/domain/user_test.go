package domain

import "testing"

func TestNormalizeEmail(t *testing.T) {
	email, err := NormalizeEmail("  User@Example.COM ")
	if err != nil {
		t.Fatalf("NormalizeEmail() error = %v", err)
	}
	if email != "user@example.com" {
		t.Fatalf("NormalizeEmail() = %q, want %q", email, "user@example.com")
	}
}

func TestNormalizeEmailRejectsDisplayAddress(t *testing.T) {
	if _, err := NormalizeEmail("User <user@example.com>"); err == nil {
		t.Fatal("expected display address to be rejected")
	}
}

func TestValidatePassword(t *testing.T) {
	if err := ValidatePassword("short"); err == nil {
		t.Fatal("expected short password to be rejected")
	}
	if err := ValidatePassword("valid-password"); err != nil {
		t.Fatalf("ValidatePassword() error = %v", err)
	}
}
