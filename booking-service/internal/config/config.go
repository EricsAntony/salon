package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/viper"
	"github.com/subosito/gotenv"
)

// Config holds all configuration for the booking service
type Config struct {
	Environment      string `mapstructure:"environment"`
	Port             int    `mapstructure:"port"`
	DatabaseURL      string `mapstructure:"database_url"`
	JWTAccessSecret  string `mapstructure:"jwt_access_secret"`
	JWTRefreshSecret string `mapstructure:"jwt_refresh_secret"`
	
	// External service URLs
	UserServiceURL  string `mapstructure:"user_service_url"`
	SalonServiceURL string `mapstructure:"salon_service_url"`
	
	// Default configuration values
	DefaultBufferTimeMinutes       int     `mapstructure:"default_buffer_time_minutes"`
	DefaultCancellationCutoffHours int     `mapstructure:"default_cancellation_cutoff_hours"`
	DefaultRescheduleWindowHours   int     `mapstructure:"default_reschedule_window_hours"`
	DefaultMaxAdvanceBookingDays   int     `mapstructure:"default_max_advance_booking_days"`
	DefaultBookingFeeAmount        float64 `mapstructure:"default_booking_fee_amount"`
	DefaultGSTPercentage           float64 `mapstructure:"default_gst_percentage"`
}

// Load loads configuration from environment variables and config files
func Load() (*Config, error) {
	// Load .env file if it exists
	_ = gotenv.Load()

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("../configs")
	viper.AddConfigPath("../../configs")

	// Set defaults
	setDefaults()

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Override with environment variables
	viper.AutomaticEnv()

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Override with environment-specific variables
	overrideWithEnv(&config)

	// Validate required fields
	if err := validate(&config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

func setDefaults() {
	viper.SetDefault("environment", "development")
	viper.SetDefault("port", 8082)
	viper.SetDefault("user_service_url", "http://localhost:8080")
	viper.SetDefault("salon_service_url", "http://localhost:8081")
	
	// Default booking configuration
	viper.SetDefault("default_buffer_time_minutes", 15)
	viper.SetDefault("default_cancellation_cutoff_hours", 2)
	viper.SetDefault("default_reschedule_window_hours", 4)
	viper.SetDefault("default_max_advance_booking_days", 30)
	viper.SetDefault("default_booking_fee_amount", 50.0)
	viper.SetDefault("default_gst_percentage", 18.0)
}

func overrideWithEnv(config *Config) {
	if env := os.Getenv("BOOKING_SERVICE_ENV"); env != "" {
		config.Environment = env
	}
	
	if port := os.Getenv("BOOKING_SERVICE_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Port = p
		}
	}
	
	if dbURL := os.Getenv("BOOKING_SERVICE_DB_URL"); dbURL != "" {
		config.DatabaseURL = dbURL
	}
	
	if secret := os.Getenv("BOOKING_SERVICE_JWT_ACCESSSECRET"); secret != "" {
		config.JWTAccessSecret = secret
	}
	
	if secret := os.Getenv("BOOKING_SERVICE_JWT_REFRESHSECRET"); secret != "" {
		config.JWTRefreshSecret = secret
	}
	
	if url := os.Getenv("USER_SERVICE_URL"); url != "" {
		config.UserServiceURL = url
	}
	
	if url := os.Getenv("SALON_SERVICE_URL"); url != "" {
		config.SalonServiceURL = url
	}
}

func validate(config *Config) error {
	if config.DatabaseURL == "" {
		return fmt.Errorf("database_url is required")
	}
	
	if config.JWTAccessSecret == "" {
		return fmt.Errorf("jwt_access_secret is required")
	}
	
	if config.JWTRefreshSecret == "" {
		return fmt.Errorf("jwt_refresh_secret is required")
	}
	
	if config.Port <= 0 || config.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}
	
	return nil
}
