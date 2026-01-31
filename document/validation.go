package document

import (
	"fmt"
	"strings"
)

// Validatable is an interface that documents can implement to provide
// custom validation logic. When implemented, the Validate method is
// called automatically before insert and replace operations.
//
// If validation fails, return an error (preferably a ValidationError
// or MultiValidationError for structured error reporting).
//
// Example:
//
//	type User struct {
//	    document.Base `bson:",inline"`
//	    Email    string `bson:"email"`
//	    Age      int    `bson:"age"`
//	    Username string `bson:"username"`
//	}
//
//	func (u *User) Validate() error {
//	    var errs document.MultiValidationError
//
//	    if u.Email == "" {
//	        errs = append(errs, document.ValidationError{
//	            Field:   "email",
//	            Message: "email is required",
//	        })
//	    }
//
//	    if u.Age < 0 || u.Age > 150 {
//	        errs = append(errs, document.ValidationError{
//	            Field:   "age",
//	            Message: "age must be between 0 and 150",
//	        })
//	    }
//
//	    if len(u.Username) < 3 {
//	        errs = append(errs, document.ValidationError{
//	            Field:   "username",
//	            Message: "username must be at least 3 characters",
//	        })
//	    }
//
//	    if len(errs) > 0 {
//	        return errs
//	    }
//	    return nil
//	}
type Validatable interface {
	// Validate checks the document's data and returns an error if invalid.
	// Called automatically before InsertOne and ReplaceOne operations.
	Validate() error
}

// ValidationError represents a validation error for a specific field.
// Use this for structured error reporting that can be easily parsed
// by API handlers to return field-specific error messages.
//
// Example:
//
//	return document.ValidationError{
//	    Field:   "email",
//	    Message: "invalid email format",
//	    Value:   user.Email,
//	}
type ValidationError struct {
	// Field is the name of the field that failed validation.
	Field string

	// Message describes why validation failed.
	Message string

	// Value is the invalid value (optional, for debugging).
	Value any
}

// Error implements the error interface.
func (e ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("%s: %s", e.Field, e.Message)
	}
	return e.Message
}

// MultiValidationError is a collection of validation errors.
// Use this when multiple fields fail validation simultaneously.
//
// Example:
//
//	func (u *User) Validate() error {
//	    var errs document.MultiValidationError
//
//	    if u.Email == "" {
//	        errs = append(errs, document.ValidationError{Field: "email", Message: "required"})
//	    }
//	    if u.Name == "" {
//	        errs = append(errs, document.ValidationError{Field: "name", Message: "required"})
//	    }
//
//	    if len(errs) > 0 {
//	        return errs
//	    }
//	    return nil
//	}
type MultiValidationError []ValidationError

// Error implements the error interface.
// Returns a combined message with all validation errors.
func (m MultiValidationError) Error() string {
	if len(m) == 0 {
		return "validation failed"
	}
	if len(m) == 1 {
		return m[0].Error()
	}

	var sb strings.Builder
	sb.WriteString("validation failed: ")
	for i, err := range m {
		if i > 0 {
			sb.WriteString("; ")
		}
		sb.WriteString(err.Error())
	}
	return sb.String()
}

// Fields returns a map of field names to error messages.
// Useful for API responses that need field-specific errors.
//
// Example:
//
//	if err != nil {
//	    if mve, ok := err.(document.MultiValidationError); ok {
//	        // Returns: {"email": "required", "name": "required"}
//	        fieldErrors := mve.Fields()
//	        return c.JSON(400, fieldErrors)
//	    }
//	}
func (m MultiValidationError) Fields() map[string]string {
	result := make(map[string]string, len(m))
	for _, err := range m {
		if err.Field != "" {
			result[err.Field] = err.Message
		}
	}
	return result
}

// HasField checks if there's a validation error for the given field.
func (m MultiValidationError) HasField(field string) bool {
	for _, err := range m {
		if err.Field == field {
			return true
		}
	}
	return false
}

// FieldError returns the first validation error for the given field,
// or nil if there's no error for that field.
func (m MultiValidationError) FieldError(field string) *ValidationError {
	for _, err := range m {
		if err.Field == field {
			return &err
		}
	}
	return nil
}

// NewValidationError creates a new ValidationError with the given field and message.
func NewValidationError(field, message string) ValidationError {
	return ValidationError{
		Field:   field,
		Message: message,
	}
}

// NewValidationErrorWithValue creates a new ValidationError with field, message, and the invalid value.
func NewValidationErrorWithValue(field, message string, value any) ValidationError {
	return ValidationError{
		Field:   field,
		Message: message,
		Value:   value,
	}
}
