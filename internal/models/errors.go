package models

import "errors"

// Domain-level sentinel errors. Layers wrap these with fmt.Errorf("%w: ...") so
// the HTTP layer can pick a status code via errors.Is instead of matching text.
var (
	// ErrNotFound is returned when a subscription does not exist.
	ErrNotFound = errors.New("subscription not found")
	// ErrValidation is returned when input fails validation.
	ErrValidation = errors.New("validation error")
)
