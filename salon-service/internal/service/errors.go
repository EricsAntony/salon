package service

import "fmt"

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	return "validation failed"
}

func NewValidationError(field, message string) ValidationErrors {
	return ValidationErrors{{Field: field, Message: message}}
}

func AppendValidationError(errs ValidationErrors, field, message string) ValidationErrors {
	return append(errs, ValidationError{Field: field, Message: message})
}

type ConflictError struct {
	Resource string
	Detail   string
}

func (c ConflictError) Error() string {
	if c.Detail != "" {
		return fmt.Sprintf("%s conflict: %s", c.Resource, c.Detail)
	}
	return fmt.Sprintf("%s conflict", c.Resource)
}
