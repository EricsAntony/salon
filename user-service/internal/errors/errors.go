package errors

import (
	"errors"
	"net/http"
)

// Application error types
var (
	ErrInvalidInput      = errors.New("invalid input")
	ErrUserNotFound      = errors.New("user not found")
	ErrUserExists        = errors.New("user already exists")
	ErrInvalidOTP        = errors.New("invalid or expired OTP")
	ErrOTPExpired        = errors.New("OTP expired")
	ErrOTPNotRequested   = errors.New("OTP not requested")
	ErrUnauthorized      = errors.New("unauthorized")
	ErrForbidden         = errors.New("forbidden")
	ErrRateLimited       = errors.New("rate limited")
	ErrInternalError     = errors.New("internal server error")
	ErrUserNotRegistered = errors.New("user not registered")
)

const (
	OtpExpired        = "OTP_EXPIRED"
	UserNotRegistered = "USER_NOT_REGISTERED"
)

// APIError represents an API error with HTTP status code
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Type    string `json:"type,omitempty"`
}

func (e APIError) Error() string {
	return e.Message
}

// NewAPIError creates a new API error
func NewAPIError(code int, message, errorType string) *APIError {
	return &APIError{
		Code:    code,
		Message: message,
		Type:    errorType,
	}
}

// MapToAPIError maps internal errors to API errors
func MapToAPIError(err error) *APIError {
	switch {
	case errors.Is(err, ErrInvalidInput):
		return NewAPIError(http.StatusBadRequest, "Invalid input provided", "validation_error")
	case errors.Is(err, ErrUserNotFound):
		return NewAPIError(http.StatusNotFound, "User not found", "not_found")
	case errors.Is(err, ErrUserExists):
		return NewAPIError(http.StatusConflict, "User already exists", "conflict")
	case errors.Is(err, ErrInvalidOTP), errors.Is(err, ErrOTPNotRequested):
		return NewAPIError(http.StatusUnauthorized, "Invalid or expired OTP", "auth_error")
	case errors.Is(err, ErrUnauthorized):
		return NewAPIError(http.StatusUnauthorized, "Authentication required", "auth_error")
	case errors.Is(err, ErrForbidden):
		return NewAPIError(http.StatusForbidden, "Access denied", "auth_error")
	case errors.Is(err, ErrRateLimited):
		return NewAPIError(http.StatusTooManyRequests, "Rate limit exceeded", "rate_limit")
	case errors.Is(err, ErrOTPExpired):
		return NewAPIError(http.StatusUnauthorized, "Invalid or expired OTP", OtpExpired)
	case errors.Is(err, ErrUserNotRegistered):
		return NewAPIError(http.StatusNotFound, "User not found", UserNotRegistered)
	default:
		return NewAPIError(http.StatusInternalServerError, "Internal server error", "internal_error")
	}
}
