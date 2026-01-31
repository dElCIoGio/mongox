package repository

import (
	"errors"
	"fmt"
)

// Sentinel errors for common repository operations.
var (
	// ErrNotFound is returned when a document matching the query is not found.
	ErrNotFound = errors.New("repository: document not found")

	// ErrDuplicateKey is returned when an insert or update violates a unique index constraint.
	ErrDuplicateKey = errors.New("repository: duplicate key error")

	// ErrInvalidFilter is returned when the provided filter is invalid or malformed.
	ErrInvalidFilter = errors.New("repository: invalid filter")

	// ErrValidation is returned when document validation fails.
	ErrValidation = errors.New("repository: validation failed")

	// ErrNilDocument is returned when a nil document is passed to an operation that requires one.
	ErrNilDocument = errors.New("repository: nil document")

	// ErrNilUpdate is returned when a nil update is passed to an update operation.
	ErrNilUpdate = errors.New("repository: nil update")
)

// ValidationError represents a validation error for a specific field.
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error on field %q: %s", e.Field, e.Message)
}

// Unwrap returns ErrValidation to allow errors.Is(err, ErrValidation) to work.
func (e ValidationError) Unwrap() error {
	return ErrValidation
}

// ValidationErrors represents multiple validation errors.
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return "validation failed"
	}
	if len(e) == 1 {
		return e[0].Error()
	}
	return fmt.Sprintf("validation failed: %d errors (first: %s)", len(e), e[0].Message)
}

// Unwrap returns ErrValidation to allow errors.Is(err, ErrValidation) to work.
func (e ValidationErrors) Unwrap() error {
	return ErrValidation
}

// NewValidationError creates a new ValidationError for the given field and message.
func NewValidationError(field, message string) ValidationError {
	return ValidationError{Field: field, Message: message}
}
