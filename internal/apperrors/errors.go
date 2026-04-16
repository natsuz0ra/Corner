package apperrors

import "errors"

// Domain sentinel errors for repositories; services/controllers match on these.
var (
	// ErrNotFound means the requested resource does not exist.
	ErrNotFound = errors.New("not found")
	// ErrInvalidInput means the input parameters are invalid.
	ErrInvalidInput = errors.New("invalid input")
)
