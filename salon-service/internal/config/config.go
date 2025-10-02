package config

import (
	"fmt"
	"strings"

	sharedConfig "github.com/EricsAntony/salon/salon-shared/config"
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
	Log struct {
		Level       string
		ServiceName string
	}
	JWT struct {
		AccessSecret     string
		RefreshSecret    string
		AccessTTLMinutes int
		RefreshTTLDays   int
	}
}

func Load() (*Config, error) {
	_ = gotenv.Load()
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./configs")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.SetEnvPrefix("SALON_SERVICE")
	v.AutomaticEnv()

	v.SetDefault("env", "dev")
	v.SetDefault("server.port", "8081")
	v.SetDefault("log.level", "info")
	v.SetDefault("log.servicename", "salon-service")
	v.SetDefault("jwt.accessttlminutes", 15)
	v.SetDefault("jwt.refreshttldays", 7)

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("read config: %w", err)
		}
	}

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return cfg, nil
}

func (c *Config) ToShared() *sharedConfig.Config {
	return &sharedConfig.Config{
		Env: c.Env,
		Server: struct{ Port string }{Port: c.Server.Port},
		DB: struct{ URL string }{URL: c.DB.URL},
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
		Log: struct {
			Level       string
			ServiceName string
		}{
			Level:       c.Log.Level,
			ServiceName: c.Log.ServiceName,
		},
	}
}
