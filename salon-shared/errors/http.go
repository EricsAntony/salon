package errors

import (
	"encoding/json"
	"errors"
	"net/http"
)

// APIError represents an HTTP API error response
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

// MapToAPIError maps internal errors to HTTP API errors
func MapToAPIError(err error) *APIError {
	// Handle validation errors
	if _, ok := err.(ValidationErrors); ok {
		return &APIError{
			Code:    http.StatusBadRequest,
			Message: "Validation failed",
			Type:    ErrorTypeValidation,
		}
	}
	
	// Handle specific error types
	switch {
	case errors.Is(err, ErrUnauthorized):
		return NewAPIError(http.StatusUnauthorized, "Authentication required", ErrorTypeAuth)
	case errors.Is(err, ErrForbidden):
		return NewAPIError(http.StatusForbidden, "Access denied", ErrorTypeAuth)
	case errors.Is(err, ErrInvalidToken):
		return NewAPIError(http.StatusUnauthorized, "Invalid token", ErrorTypeAuth)
	case errors.Is(err, ErrTokenExpired):
		return NewAPIError(http.StatusUnauthorized, "Token expired", ErrorTypeAuth)
	case errors.Is(err, ErrInvalidCredentials):
		return NewAPIError(http.StatusUnauthorized, "Invalid credentials", ErrorTypeAuth)
		
	case errors.Is(err, ErrUserNotFound):
		return NewAPIError(http.StatusNotFound, "User not found", ErrorTypeNotFound)
	case errors.Is(err, ErrStaffNotFound):
		return NewAPIError(http.StatusNotFound, "Staff not found", ErrorTypeNotFound)
	case errors.Is(err, ErrNotFound):
		return NewAPIError(http.StatusNotFound, "Resource not found", ErrorTypeNotFound)
		
	case errors.Is(err, ErrUserExists):
		return NewAPIError(http.StatusConflict, "User already exists", ErrorTypeConflict)
	case errors.Is(err, ErrStaffExists):
		return NewAPIError(http.StatusConflict, "Staff already exists", ErrorTypeConflict)
	case errors.Is(err, ErrAlreadyExists):
		return NewAPIError(http.StatusConflict, "Resource already exists", ErrorTypeConflict)
		
	case errors.Is(err, ErrInvalidOTP):
		return NewAPIError(http.StatusUnauthorized, "Invalid OTP", ErrorTypeAuth)
	case errors.Is(err, ErrOTPExpired):
		return NewAPIError(http.StatusUnauthorized, "OTP expired", CodeOTPExpired)
	case errors.Is(err, ErrOTPNotRequested):
		return NewAPIError(http.StatusBadRequest, "OTP not requested", ErrorTypeValidation)
	case errors.Is(err, ErrOTPAttemptsExceeded):
		return NewAPIError(http.StatusTooManyRequests, "Too many OTP attempts", ErrorTypeRateLimit)
		
	case errors.Is(err, ErrRateLimited):
		return NewAPIError(http.StatusTooManyRequests, "Rate limit exceeded", ErrorTypeRateLimit)
		
	case errors.Is(err, ErrInvalidInput):
		return NewAPIError(http.StatusBadRequest, "Invalid input", ErrorTypeValidation)
		
	case errors.Is(err, ErrServiceUnavailable):
		return NewAPIError(http.StatusServiceUnavailable, "Service unavailable", ErrorTypeUnavailable)
	case errors.Is(err, ErrDatabaseError):
		return NewAPIError(http.StatusInternalServerError, "Database error", ErrorTypeInternal)
	}
	
	// Handle custom error types
	if conflictErr, ok := err.(*ConflictError); ok {
		return NewAPIError(http.StatusConflict, conflictErr.Error(), ErrorTypeConflict)
	}
	
	if notFoundErr, ok := err.(*NotFoundError); ok {
		return NewAPIError(http.StatusNotFound, notFoundErr.Error(), ErrorTypeNotFound)
	}
	
	if authErr, ok := err.(*AuthError); ok {
		return NewAPIError(http.StatusUnauthorized, authErr.Error(), ErrorTypeAuth)
	}
	
	// Default to internal server error
	return NewAPIError(http.StatusInternalServerError, "Internal server error", ErrorTypeInternal)
}

// WriteAPIError writes an API error response to the HTTP response writer
func WriteAPIError(w http.ResponseWriter, err error) {
	apiErr := MapToAPIError(err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(apiErr.Code)
	
	response := map[string]interface{}{
		"success": false,
		"error": map[string]interface{}{
			"code":    apiErr.Code,
			"message": apiErr.Message,
			"type":    apiErr.Type,
		},
	}
	
	// Add validation details if applicable
	if validationErrs, ok := err.(ValidationErrors); ok {
		response["error"].(map[string]interface{})["details"] = validationErrs
	}
	
	json.NewEncoder(w).Encode(response)
}
