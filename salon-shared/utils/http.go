package utils

import (
	"encoding/json"
	"net/http"
	"strings"
)

// WriteJSON writes a JSON response with the given status code
func WriteJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

// WriteError writes a standardized error response
func WriteError(w http.ResponseWriter, status int, message string) {
	WriteJSON(w, status, map[string]any{
		"success": false,
		"error": map[string]any{
			"code":    status,
			"type":    http.StatusText(status),
			"message": message,
		},
	})
}

// WriteSuccess writes a standardized success response
func WriteSuccess(w http.ResponseWriter, data any) {
	WriteJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    data,
	})
}

// GetClientIP extracts the client IP from the request
func GetClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP in case of multiple
		return strings.Split(xff, ",")[0]
	}
	
	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	
	// Fall back to RemoteAddr
	return r.RemoteAddr
}
