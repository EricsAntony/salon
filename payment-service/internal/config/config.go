package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all configuration for the payment service
type Config struct {
	Environment    string
	Port           string
	LogLevel       string
	DatabaseURL    string
	MigrationsPath string

	// Payment Gateway Configurations
	StripeSecretKey      string
	StripeWebhookSecret  string
	RazorpayKeyID        string
	RazorpayKeySecret    string
	RazorpayWebhookSecret string

	// Service Configuration
	DefaultCurrency      string
	PaymentTimeoutMinutes int
	MaxRetryAttempts     int
	IdempotencyTTLHours  int

	// External Service URLs
	BookingServiceURL    string
	NotificationServiceURL string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Environment:    getEnv("PAYMENT_SERVICE_ENV", "dev"),
		Port:           getEnv("PAYMENT_SERVICE_PORT", "8082"),
		LogLevel:       getEnv("PAYMENT_SERVICE_LOG_LEVEL", "info"),
		DatabaseURL:    getEnv("PAYMENT_SERVICE_DB_URL", "postgres://postgres:password@localhost:5432/salon?sslmode=disable"),
		MigrationsPath: getEnv("PAYMENT_SERVICE_MIGRATIONS_PATH", "./migrations"),

		// Payment Gateways
		StripeSecretKey:       getEnv("STRIPE_SECRET_KEY", ""),
		StripeWebhookSecret:   getEnv("STRIPE_WEBHOOK_SECRET", ""),
		RazorpayKeyID:         getEnv("RAZORPAY_KEY_ID", ""),
		RazorpayKeySecret:     getEnv("RAZORPAY_KEY_SECRET", ""),
		RazorpayWebhookSecret: getEnv("RAZORPAY_WEBHOOK_SECRET", ""),

		// Service Configuration
		DefaultCurrency:      getEnv("DEFAULT_CURRENCY", "INR"),
		PaymentTimeoutMinutes: getEnvInt("PAYMENT_TIMEOUT_MINUTES", 15),
		MaxRetryAttempts:     getEnvInt("MAX_RETRY_ATTEMPTS", 3),
		IdempotencyTTLHours:  getEnvInt("IDEMPOTENCY_TTL_HOURS", 24),

		// External Services
		BookingServiceURL:     getEnv("BOOKING_SERVICE_URL", "http://localhost:8083"),
		NotificationServiceURL: getEnv("NOTIFICATION_SERVICE_URL", "http://localhost:8084"),
	}

	// Validate required configuration
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("PAYMENT_SERVICE_DB_URL is required")
	}

	return cfg, nil
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt gets an integer environment variable with a default value
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
