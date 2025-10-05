package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/EricsAntony/salon/salon-shared/auth"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

// AuditConfig configures the audit middleware for different services
type AuditConfig struct {
	ServiceName       string
	SensitivePatterns []string
	MaxBodySize       int
}

// DefaultAuditConfig returns default configuration
func DefaultAuditConfig(serviceName string) AuditConfig {
	return AuditConfig{
		ServiceName: serviceName,
		SensitivePatterns: []string{
			"/otp/request",
			"/user/register",
			"/user/authenticate",
			"/staff/otp",
			"/staff/auth",
			"/auth/refresh",
			"/auth/revoke",
			"/users/",
			"/salons",
			"/branches",
			"/categories",
			"/services",
			"/staff",
		},
		MaxBodySize: 1024,
	}
}

// AuditEvent represents an audit log entry
type AuditEvent struct {
	Timestamp   time.Time `json:"timestamp"`
	ServiceName string    `json:"service_name"`
	UserID      string    `json:"user_id,omitempty"`
	Action      string    `json:"action"`
	Resource    string    `json:"resource"`
	ResourceID  string    `json:"resource_id,omitempty"`
	Method      string    `json:"method"`
	Path        string    `json:"path"`
	ClientIP    string    `json:"client_ip"`
	UserAgent   string    `json:"user_agent,omitempty"`
	StatusCode  int       `json:"status_code"`
	Success     bool      `json:"success"`
	Details     any       `json:"details,omitempty"`
}

// ResponseWriter wraps http.ResponseWriter to capture status code
type ResponseWriter struct {
	http.ResponseWriter
	statusCode int
	body       bytes.Buffer
}

func (rw *ResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *ResponseWriter) Write(b []byte) (int, error) {
	rw.body.Write(b)
	return rw.ResponseWriter.Write(b)
}

// AuditMiddleware creates an audit logging middleware with the given configuration
func AuditMiddleware(config AuditConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only audit sensitive operations
			if !isSensitiveOperation(r, config.SensitivePatterns) {
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()
			
			// Capture request body for POST/PUT operations
			var requestBody []byte
			if r.Method == "POST" || r.Method == "PUT" {
				if r.Body != nil {
					requestBody, _ = io.ReadAll(r.Body)
					r.Body = io.NopCloser(bytes.NewBuffer(requestBody))
				}
			}

			// Wrap response writer to capture status code
			rw := &ResponseWriter{
				ResponseWriter: w,
				statusCode:     200,
			}

			next.ServeHTTP(rw, r)

			// Log audit event
			event := createAuditEvent(r, rw, start, requestBody, config)
			logAuditEvent(event)
		})
	}
}

// isSensitiveOperation determines if an operation should be audited
func isSensitiveOperation(r *http.Request, patterns []string) bool {
	path := r.URL.Path
	for _, pattern := range patterns {
		if strings.Contains(path, pattern) {
			return true
		}
	}
	return false
}

// createAuditEvent creates an audit event from the request and response
func createAuditEvent(r *http.Request, rw *ResponseWriter, start time.Time, requestBody []byte, config AuditConfig) AuditEvent {
	userID := ""
	if uid := r.Context().Value(auth.CtxUserID); uid != nil {
		userID = uid.(string)
	}

	action := determineAction(r)
	resource, resourceID := determineResource(r)

	event := AuditEvent{
		Timestamp:   start,
		ServiceName: config.ServiceName,
		UserID:      userID,
		Action:      action,
		Resource:    resource,
		ResourceID:  resourceID,
		Method:      r.Method,
		Path:        r.URL.Path,
		ClientIP:    GetClientIP(r),
		UserAgent:   r.Header.Get("User-Agent"),
		StatusCode:  rw.statusCode,
		Success:     rw.statusCode >= 200 && rw.statusCode < 400,
	}

	// Add request details for sensitive operations
	if len(requestBody) > 0 && len(requestBody) < config.MaxBodySize {
		var requestData map[string]interface{}
		if err := json.Unmarshal(requestBody, &requestData); err == nil {
			// Remove sensitive fields
			sanitizeRequestData(requestData)
			event.Details = map[string]interface{}{
				"request": requestData,
			}
		}
	}

	return event
}

// determineAction determines the action type from the request
func determineAction(r *http.Request) string {
	switch r.Method {
	case "POST":
		if strings.Contains(r.URL.Path, "/auth") {
			return "authenticate"
		}
		if strings.Contains(r.URL.Path, "/register") {
			return "register"
		}
		if strings.Contains(r.URL.Path, "/otp") {
			return "request_otp"
		}
		if strings.Contains(r.URL.Path, "/revoke") {
			return "revoke_token"
		}
		return "create"
	case "PUT":
		return "update"
	case "DELETE":
		return "delete"
	case "GET":
		return "read"
	default:
		return r.Method
	}
}

// determineResource determines the resource type and ID from the request
func determineResource(r *http.Request) (string, string) {
	path := r.URL.Path
	
	// User service resources
	if strings.Contains(path, "/users/") {
		userID := chi.URLParam(r, "id")
		return "user", userID
	}
	if strings.Contains(path, "/user/") {
		userID := chi.URLParam(r, "id")
		return "user", userID
	}
	
	// Salon service resources
	if strings.Contains(path, "/staff") {
		staffID := chi.URLParam(r, "staffID")
		return "staff", staffID
	}
	if strings.Contains(path, "/services") {
		serviceID := chi.URLParam(r, "serviceID")
		return "service", serviceID
	}
	if strings.Contains(path, "/categories") {
		categoryID := chi.URLParam(r, "categoryID")
		return "category", categoryID
	}
	if strings.Contains(path, "/branches") {
		branchID := chi.URLParam(r, "branchID")
		return "branch", branchID
	}
	if strings.Contains(path, "/salons") {
		salonID := chi.URLParam(r, "salonID")
		return "salon", salonID
	}
	
	// Common resources
	if strings.Contains(path, "/otp") {
		return "otp", ""
	}
	if strings.Contains(path, "/auth") {
		return "auth", ""
	}
	
	return "unknown", ""
}

// sanitizeRequestData removes sensitive information from request data
func sanitizeRequestData(data map[string]interface{}) {
	sensitiveFields := []string{"password", "otp", "token", "secret", "refresh_token"}
	
	for _, field := range sensitiveFields {
		if _, exists := data[field]; exists {
			data[field] = "[REDACTED]"
		}
	}
}

// logAuditEvent logs the audit event
func logAuditEvent(event AuditEvent) {
	logger := log.Info()
	
	if !event.Success {
		logger = log.Warn()
	}
	
	logger.
		Str("audit_type", "operation").
		Str("service_name", event.ServiceName).
		Str("user_id", event.UserID).
		Str("action", event.Action).
		Str("resource", event.Resource).
		Str("resource_id", event.ResourceID).
		Str("method", event.Method).
		Str("path", event.Path).
		Str("client_ip", event.ClientIP).
		Int("status_code", event.StatusCode).
		Bool("success", event.Success).
		Interface("details", event.Details).
		Msg("audit log")
}

// GetClientIP extracts the client IP from the request
func GetClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return strings.Split(xff, ",")[0]
	}
	
	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	
	// Fall back to RemoteAddr
	return r.RemoteAddr
}
