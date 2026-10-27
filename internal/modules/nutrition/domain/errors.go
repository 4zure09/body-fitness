package domain

import "errors"

var (
	ErrConflict         = errors.New("nutrition catalog item already exists")
	ErrInvalidInput     = errors.New("invalid nutrition catalog input")
	ErrInvalidReference = errors.New("referenced nutrition item does not exist")
	ErrNotFound         = errors.New("nutrition catalog item not found")
)
