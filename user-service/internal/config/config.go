package config

import (
	"fmt"
	"strings"

	"github.com/subosito/gotenv"
	"github.com/spf13/viper"
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
	v.SetDefault("ratelimit.otprequestsperminute", 3)
	v.SetDefault("log.level", "info")
	v.SetDefault("log.servicename", "user-service")

	_ = v.ReadInConfig() // optional

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("config unmarshal: %w", err)
	}
	return cfg, nil
}
