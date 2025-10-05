package errors

import (
	"errors"
	"fmt"
)

// Common application errors
var (
	// Authentication errors
	ErrUnauthorized      = errors.New("unauthorized")
	ErrForbidden         = errors.New("forbidden")
	ErrInvalidToken      = errors.New("invalid token")
	ErrTokenExpired      = errors.New("token expired")
	ErrInvalidCredentials = errors.New("invalid credentials")
	
	// User/Staff errors
	ErrUserNotFound      = errors.New("user not found")
	ErrUserExists        = errors.New("user already exists")
	ErrStaffNotFound     = errors.New("staff not found")
	ErrStaffExists       = errors.New("staff already exists")
	
	// OTP errors
	ErrInvalidOTP        = errors.New("invalid OTP")
	ErrOTPExpired        = errors.New("OTP expired")
	ErrOTPNotRequested   = errors.New("OTP not requested")
	ErrOTPAttemptsExceeded = errors.New("OTP attempts exceeded")
	
	// Rate limiting errors
	ErrRateLimited       = errors.New("rate limit exceeded")
	
	// Resource errors
	ErrNotFound          = errors.New("resource not found")
	ErrAlreadyExists     = errors.New("resource already exists")
	ErrInvalidInput      = errors.New("invalid input")
	
	// System errors
	ErrInternalError     = errors.New("internal server error")
	ErrServiceUnavailable = errors.New("service unavailable")
	ErrDatabaseError     = errors.New("database error")
)

// ConflictError represents a resource conflict error
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

// NewConflictError creates a new conflict error
func NewConflictError(resource, detail string) *ConflictError {
	return &ConflictError{
		Resource: resource,
		Detail:   detail,
	}
}

// NotFoundError represents a resource not found error
type NotFoundError struct {
	Resource string
	ID       string
}

func (n NotFoundError) Error() string {
	if n.ID != "" {
		return fmt.Sprintf("%s with ID '%s' not found", n.Resource, n.ID)
	}
	return fmt.Sprintf("%s not found", n.Resource)
}

// NewNotFoundError creates a new not found error
func NewNotFoundError(resource, id string) *NotFoundError {
	return &NotFoundError{
		Resource: resource,
		ID:       id,
	}
}

// AuthError represents authentication/authorization errors
type AuthError struct {
	Type    string
	Message string
}

func (a AuthError) Error() string {
	return fmt.Sprintf("%s: %s", a.Type, a.Message)
}

// NewAuthError creates a new authentication error
func NewAuthError(errorType, message string) *AuthError {
	return &AuthError{
		Type:    errorType,
		Message: message,
	}
}

// Error type constants
const (
	ErrorTypeValidation    = "validation_error"
	ErrorTypeAuth          = "auth_error"
	ErrorTypeNotFound      = "not_found"
	ErrorTypeConflict      = "conflict"
	ErrorTypeRateLimit     = "rate_limit"
	ErrorTypeInternal      = "internal_error"
	ErrorTypeUnavailable   = "service_unavailable"
)

// Error code constants
const (
	CodeOTPExpired        = "OTP_EXPIRED"
	CodeUserNotRegistered = "USER_NOT_REGISTERED"
	CodeStaffNotFound     = "STAFF_NOT_FOUND"
	CodeInvalidUserType   = "INVALID_USER_TYPE"
	CodeAccessDenied      = "ACCESS_DENIED"
)
