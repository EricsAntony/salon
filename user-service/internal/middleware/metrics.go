package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"user-service/internal/metrics"
)

// MetricsMiddleware records HTTP request metrics
func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Create a response writer wrapper to capture status code
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		
		// Get route pattern for better grouping
		rctx := chi.RouteContext(r.Context())
		endpoint := r.URL.Path
		if rctx != nil && rctx.RoutePattern() != "" {
			endpoint = rctx.RoutePattern()
		}
		
		next.ServeHTTP(ww, r)
		
		duration := time.Since(start)
		statusCode := strconv.Itoa(ww.Status())
		
		metrics.RecordHTTPRequest(r.Method, endpoint, statusCode, duration)
	})
}
