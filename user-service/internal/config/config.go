package config

import (
	"fmt"
	"strings"

	"github.com/subosito/gotenv"
	"github.com/spf13/viper"
	sharedConfig "github.com/EricsAntony/salon/salon-shared/config"
)

type Config struct {
	Env    string
	Server struct {
		Port string
	}
	DB struct {
		URL string
	}
	JWT struct {
		AccessSecret     string
		RefreshSecret    string
		AccessTTLMinutes int
		RefreshTTLDays   int
	}
	OTP struct {
		ExpiryMinutes int
		MaxFailedAttempts int
		FailureWindowMinutes int
	}
	RateLimit struct {
		OTPRequestsPerMinute int
	}
	Log struct {
		Level       string
		ServiceName string
	}
}

func Load() (*Config, error) {
	// Load .env if present
	_ = gotenv.Load()
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./configs")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.SetEnvPrefix("USER_SERVICE")

	// Defaults
	v.SetDefault("server.port", "8080")
	v.SetDefault("env", "dev")
	v.SetDefault("jwt.accessttlminutes", 15)
	v.SetDefault("jwt.refreshttldays", 7)
	v.SetDefault("otp.expiryminutes", 5)
	v.SetDefault("otp.maxfailedattempts", 3)
	v.SetDefault("otp.failurewindowminutes", 15)
	v.SetDefault("ratelimit.otprequestsperminute", 3)
	v.SetDefault("log.level", "info")
	v.SetDefault("log.servicename", "user-service")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

// ToSharedConfig converts local config to shared config for compatibility
func (c *Config) ToSharedConfig() *sharedConfig.Config {
	return &sharedConfig.Config{
		Env: c.Env,
		Server: struct {
			Port string
		}{
			Port: c.Server.Port,
		},
		DB: struct {
			URL string
		}{
			URL: c.DB.URL,
		},
		JWT: struct {
			AccessSecret     string
			RefreshSecret    string
			AccessTTLMinutes int
			RefreshTTLDays   int
		}{
			AccessSecret:     c.JWT.AccessSecret,
			RefreshSecret:    c.JWT.RefreshSecret,
			AccessTTLMinutes: c.JWT.AccessTTLMinutes,
			RefreshTTLDays:   c.JWT.RefreshTTLDays,
		},
		OTP: struct {
			ExpiryMinutes int
		}{
			ExpiryMinutes: c.OTP.ExpiryMinutes,
		},
		RateLimit: struct {
			OTPRequestsPerMinute int
		}{
			OTPRequestsPerMinute: c.RateLimit.OTPRequestsPerMinute,
		},
		Log: struct {
			Level       string
			ServiceName string
		}{
			Level:       c.Log.Level,
			ServiceName: c.Log.ServiceName,
		},
	}
}
