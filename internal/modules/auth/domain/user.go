package domain

import (
	"fmt"
	"net/mail"
	"strings"
	"time"
)

type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

type UserStatus string

const (
	UserStatusActive  UserStatus = "active"
	UserStatusBlocked UserStatus = "blocked"
	UserStatusPending UserStatus = "pending"
)

type User struct {
	ID           string
	Email        string
	PasswordHash string
	DisplayName  string
	Status       UserStatus
	Roles        []Role
	CreatedAt    time.Time
}

func (u User) IsActive() bool {
	return u.Status == UserStatusActive
}

func NormalizeEmail(value string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if normalized == "" || len(normalized) > 254 {
		return "", fmt.Errorf("%w: email is required and must not exceed 254 characters", ErrInvalidInput)
	}
	address, err := mail.ParseAddress(normalized)
	if err != nil || address.Address != normalized {
		return "", fmt.Errorf("%w: email is invalid", ErrInvalidInput)
	}
	return normalized, nil
}

func ValidatePassword(password string) error {
	if len(password) < 8 || len(password) > 128 {
		return fmt.Errorf("%w: password must contain between 8 and 128 characters", ErrInvalidInput)
	}
	return nil
}

func NormalizeDisplayName(value string) (string, error) {
	normalized := strings.TrimSpace(value)
	if len(normalized) < 2 || len(normalized) > 80 {
		return "", fmt.Errorf("%w: display_name must contain between 2 and 80 characters", ErrInvalidInput)
	}
	return normalized, nil
}
