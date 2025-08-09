package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
	"github.com/subosito/gotenv"
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

// Load loads configuration from configs/config.yaml (optional) and env variables.
// For backward compatibility, it uses the USER_SERVICE_ prefix by default.
func Load() (*Config, error) {
	return LoadWithPrefix("USER_SERVICE", "./configs")
}

// LoadWithPrefix allows services to choose their own env prefix and config dir.
func LoadWithPrefix(prefix, configDir string) (*Config, error) {
	_ = gotenv.Load()
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(configDir)
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	if prefix != "" {
		v.SetEnvPrefix(prefix)
	}

	v.SetDefault("server.port", "8080")
	v.SetDefault("env", "dev")
	v.SetDefault("jwt.accessttlminutes", 15)
	v.SetDefault("jwt.refreshttldays", 7)
	v.SetDefault("otp.expiryminutes", 5)
	v.SetDefault("ratelimit.otprequestsperminute", 3)
	v.SetDefault("log.level", "info")
	v.SetDefault("log.servicename", "service")

	_ = v.ReadInConfig() // optional

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("config unmarshal: %w", err)
	}
	return cfg, nil
}
