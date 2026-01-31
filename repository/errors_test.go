package repository_test

import (
	"errors"
	"testing"

	"github.com/dElCIoGio/mongox/repository"
)

func TestValidationErrorString(t *testing.T) {
	err := repository.NewValidationError("email", "must be a valid email address")
	expected := `validation error on field "email": must be a valid email address`

	if err.Error() != expected {
		t.Fatalf("ValidationError.Error() mismatch.\n got: %q\nwant: %q", err.Error(), expected)
	}
}

func TestValidationErrorUnwrap(t *testing.T) {
	err := repository.NewValidationError("name", "is required")

	if !errors.Is(err, repository.ErrValidation) {
		t.Fatal("expected ValidationError to unwrap to ErrValidation")
	}
}

func TestValidationErrorsString(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		var errs repository.ValidationErrors
		if errs.Error() != "validation failed" {
			t.Fatalf("expected 'validation failed', got %q", errs.Error())
		}
	})

	t.Run("single error", func(t *testing.T) {
		errs := repository.ValidationErrors{
			repository.NewValidationError("name", "is required"),
		}
		expected := `validation error on field "name": is required`
		if errs.Error() != expected {
			t.Fatalf("ValidationErrors.Error() mismatch.\n got: %q\nwant: %q", errs.Error(), expected)
		}
	})

	t.Run("multiple errors", func(t *testing.T) {
		errs := repository.ValidationErrors{
			repository.NewValidationError("name", "is required"),
			repository.NewValidationError("email", "must be valid"),
		}
		expected := "validation failed: 2 errors (first: is required)"
		if errs.Error() != expected {
			t.Fatalf("ValidationErrors.Error() mismatch.\n got: %q\nwant: %q", errs.Error(), expected)
		}
	})
}

func TestValidationErrorsUnwrap(t *testing.T) {
	errs := repository.ValidationErrors{
		repository.NewValidationError("field", "error"),
	}

	if !errors.Is(errs, repository.ErrValidation) {
		t.Fatal("expected ValidationErrors to unwrap to ErrValidation")
	}
}

func TestSentinelErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{"ErrNotFound", repository.ErrNotFound, "repository: document not found"},
		{"ErrDuplicateKey", repository.ErrDuplicateKey, "repository: duplicate key error"},
		{"ErrInvalidFilter", repository.ErrInvalidFilter, "repository: invalid filter"},
		{"ErrValidation", repository.ErrValidation, "repository: validation failed"},
		{"ErrNilDocument", repository.ErrNilDocument, "repository: nil document"},
		{"ErrNilUpdate", repository.ErrNilUpdate, "repository: nil update"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.want {
				t.Fatalf("%s.Error() mismatch.\n got: %q\nwant: %q", tt.name, tt.err.Error(), tt.want)
			}
		})
	}
}
