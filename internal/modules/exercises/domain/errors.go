package domain

import "errors"

var (
	ErrConflict         = errors.New("catalog item already exists")
	ErrInvalidInput     = errors.New("invalid exercise catalog input")
	ErrInvalidReference = errors.New("referenced catalog item does not exist")
	ErrInUse            = errors.New("catalog item is in use")
	ErrNotFound         = errors.New("exercise catalog item not found")
)
