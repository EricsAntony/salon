package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds all configuration for the notification service
type Config struct {
	Environment    string
	Port           string
	LogLevel       string
	DatabaseURL    string
	MigrationsPath string

	// Message Broker Configuration
	MessageBroker string // "kafka" or "rabbitmq"
	
	// Kafka Configuration
	KafkaBrokers []string
	KafkaTopics  []string
	KafkaGroupID string

	// RabbitMQ Configuration
	RabbitMQURL      string
	RabbitMQExchange string
	RabbitMQQueue    string

	// Email Provider Configuration
	EmailProvider     string // "sendgrid", "smtp"
	SendGridAPIKey    string
	SMTPHost          string
	SMTPPort          int
	SMTPUsername      string
	SMTPPassword      string
	SMTPFromEmail     string
	SMTPFromName      string

	// SMS Provider Configuration
	SMSProvider       string // "twilio"
	TwilioAccountSID  string
	TwilioAuthToken   string
	TwilioFromNumber  string

	// Push Notification Configuration
	PushProvider      string // "fcm", "apns"
	FCMServerKey      string
	APNSKeyID         string
	APNSTeamID        string
	APNSBundleID      string
	APNSKeyPath       string

	// Service Configuration
	MaxRetryAttempts    int
	RetryDelaySeconds   int
	NotificationTTLDays int
	BatchSize           int
	WorkerCount         int

	// External Service URLs
	BookingServiceURL string
	UserServiceURL    string
	SalonServiceURL   string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Environment:    getEnv("NOTIFICATION_SERVICE_ENV", "dev"),
		Port:           getEnv("NOTIFICATION_SERVICE_PORT", "8084"),
		LogLevel:       getEnv("NOTIFICATION_SERVICE_LOG_LEVEL", "info"),
		DatabaseURL:    getEnv("NOTIFICATION_SERVICE_DB_URL", "postgres://postgres:password@localhost:5432/salon?sslmode=disable"),
		MigrationsPath: getEnv("NOTIFICATION_SERVICE_MIGRATIONS_PATH", "./migrations"),

		// Message Broker
		MessageBroker: getEnv("MESSAGE_BROKER", "kafka"),
		
		// Kafka
		KafkaBrokers: getEnvSlice("KAFKA_BROKERS", []string{"localhost:9092"}),
		KafkaTopics:  getEnvSlice("KAFKA_TOPICS", []string{"booking-events"}),
		KafkaGroupID: getEnv("KAFKA_GROUP_ID", "notification-service"),

		// RabbitMQ
		RabbitMQURL:      getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		RabbitMQExchange: getEnv("RABBITMQ_EXCHANGE", "booking-events"),
		RabbitMQQueue:    getEnv("RABBITMQ_QUEUE", "notifications"),

		// Email
		EmailProvider:  getEnv("EMAIL_PROVIDER", "smtp"),
		SendGridAPIKey: getEnv("SENDGRID_API_KEY", ""),
		SMTPHost:       getEnv("SMTP_HOST", "localhost"),
		SMTPPort:       getEnvInt("SMTP_PORT", 587),
		SMTPUsername:   getEnv("SMTP_USERNAME", ""),
		SMTPPassword:   getEnv("SMTP_PASSWORD", ""),
		SMTPFromEmail:  getEnv("SMTP_FROM_EMAIL", "noreply@salon.com"),
		SMTPFromName:   getEnv("SMTP_FROM_NAME", "Salon Booking"),

		// SMS
		SMSProvider:      getEnv("SMS_PROVIDER", "twilio"),
		TwilioAccountSID: getEnv("TWILIO_ACCOUNT_SID", ""),
		TwilioAuthToken:  getEnv("TWILIO_AUTH_TOKEN", ""),
		TwilioFromNumber: getEnv("TWILIO_FROM_NUMBER", ""),

		// Push Notifications
		PushProvider: getEnv("PUSH_PROVIDER", "fcm"),
		FCMServerKey: getEnv("FCM_SERVER_KEY", ""),
		APNSKeyID:    getEnv("APNS_KEY_ID", ""),
		APNSTeamID:   getEnv("APNS_TEAM_ID", ""),
		APNSBundleID: getEnv("APNS_BUNDLE_ID", ""),
		APNSKeyPath:  getEnv("APNS_KEY_PATH", ""),

		// Service Configuration
		MaxRetryAttempts:    getEnvInt("MAX_RETRY_ATTEMPTS", 3),
		RetryDelaySeconds:   getEnvInt("RETRY_DELAY_SECONDS", 30),
		NotificationTTLDays: getEnvInt("NOTIFICATION_TTL_DAYS", 30),
		BatchSize:           getEnvInt("BATCH_SIZE", 100),
		WorkerCount:         getEnvInt("WORKER_COUNT", 5),

		// External Services
		BookingServiceURL: getEnv("BOOKING_SERVICE_URL", "http://localhost:8083"),
		UserServiceURL:    getEnv("USER_SERVICE_URL", "http://localhost:8080"),
		SalonServiceURL:   getEnv("SALON_SERVICE_URL", "http://localhost:8081"),
	}

	// Validate required configuration
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("NOTIFICATION_SERVICE_DB_URL is required")
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

// getEnvSlice gets a comma-separated environment variable as a slice
func getEnvSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}
