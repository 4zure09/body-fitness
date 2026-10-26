package domain

import "errors"

var (
	ErrAccountDisabled    = errors.New("account is disabled")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidInput       = errors.New("invalid input")
	ErrInvalidToken       = errors.New("invalid token")
	ErrNotFound           = errors.New("resource not found")
)
