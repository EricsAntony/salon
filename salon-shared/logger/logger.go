package logger

import (
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/your-org/salon-shared/config"
)

// Init configures global zerolog based on the provided config.
func Init(cfg *config.Config) {
	lvl, err := zerolog.ParseLevel(strings.ToLower(cfg.Log.Level))
	if err != nil {
		lvl = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(lvl)
	log.Logger = zerolog.New(os.Stdout).With().
		Timestamp().
		Str("service", cfg.Log.ServiceName).
		Logger()
}
