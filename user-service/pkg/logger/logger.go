package logger

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/EricsAntony/salon/salon-shared/config"
)

func Init(cfg *config.Config) {
	level, err := zerolog.ParseLevel(cfg.Log.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)
	
	// Check if we should use JSON format for Loki integration
	logFormat := os.Getenv("LOG_FORMAT")
	if logFormat == "json" {
		// Use JSON format for structured logging (Loki-friendly)
		log.Logger = zerolog.New(os.Stdout).With().
			Timestamp().
			Str("service", cfg.Log.ServiceName).
			Logger()
	} else {
		// Use console format for development
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).With().
			Timestamp().
			Str("service", cfg.Log.ServiceName).
			Logger()
	}
}
