package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP metrics
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "user_service_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status_code"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "user_service_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	// OTP metrics
	OTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "user_service_otp_requests_total",
			Help: "Total number of OTP requests",
		},
		[]string{"status"},
	)

	OTPVerificationTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "user_service_otp_verification_total",
			Help: "Total number of OTP verifications",
		},
		[]string{"status"},
	)

	// User metrics
	UserRegistrationsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "user_service_user_registrations_total",
			Help: "Total number of user registrations",
		},
	)

	UserAuthenticationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "user_service_user_authentications_total",
			Help: "Total number of user authentications",
		},
		[]string{"status"},
	)

	// Database metrics
	DatabaseConnectionsActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "user_service_database_connections_active",
			Help: "Number of active database connections",
		},
	)

	DatabaseQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "user_service_database_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)

	// Rate limiting metrics
	RateLimitHitsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "user_service_rate_limit_hits_total",
			Help: "Total number of rate limit hits",
		},
		[]string{"type"},
	)
)

// RecordHTTPRequest records an HTTP request metric
func RecordHTTPRequest(method, endpoint, statusCode string, duration time.Duration) {
	HTTPRequestsTotal.WithLabelValues(method, endpoint, statusCode).Inc()
	HTTPRequestDuration.WithLabelValues(method, endpoint).Observe(duration.Seconds())
}

// RecordOTPRequest records an OTP request metric
func RecordOTPRequest(status string) {
	OTPRequestsTotal.WithLabelValues(status).Inc()
}

// RecordOTPVerification records an OTP verification metric
func RecordOTPVerification(status string) {
	OTPVerificationTotal.WithLabelValues(status).Inc()
}

// RecordUserRegistration records a user registration metric
func RecordUserRegistration() {
	UserRegistrationsTotal.Inc()
}

// RecordUserAuthentication records a user authentication metric
func RecordUserAuthentication(status string) {
	UserAuthenticationsTotal.WithLabelValues(status).Inc()
}

// RecordDatabaseQuery records a database query metric
func RecordDatabaseQuery(operation string, duration time.Duration) {
	DatabaseQueryDuration.WithLabelValues(operation).Observe(duration.Seconds())
}

// RecordRateLimitHit records a rate limit hit metric
func RecordRateLimitHit(limitType string) {
	RateLimitHitsTotal.WithLabelValues(limitType).Inc()
}
