package errors

import "fmt"

// ValidationError represents a single field validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	if len(v) == 1 {
		return fmt.Sprintf("validation failed for field '%s': %s", v[0].Field, v[0].Message)
	}
	return fmt.Sprintf("validation failed for %d fields", len(v))
}

// NewValidationError creates a new validation error for a single field
func NewValidationError(field, message string) ValidationErrors {
	return ValidationErrors{{Field: field, Message: message}}
}

// AppendValidationError adds a validation error to existing errors
func AppendValidationError(errs ValidationErrors, field, message string) ValidationErrors {
	return append(errs, ValidationError{Field: field, Message: message})
}

// MergeValidationErrors combines multiple validation error sets
func MergeValidationErrors(errorSets ...ValidationErrors) ValidationErrors {
	var result ValidationErrors
	for _, errSet := range errorSets {
		result = append(result, errSet...)
	}
	return result
}

// HasField checks if validation errors contain a specific field
func (v ValidationErrors) HasField(field string) bool {
	for _, err := range v {
		if err.Field == field {
			return true
		}
	}
	return false
}

// GetFieldError returns the error message for a specific field
func (v ValidationErrors) GetFieldError(field string) string {
	for _, err := range v {
		if err.Field == field {
			return err.Message
		}
	}
	return ""
}
