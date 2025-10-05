package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"

	"user-service/internal/api"
	"user-service/internal/cleanup"
	"user-service/internal/repository"
	"user-service/internal/service"
	"user-service/internal/tracing"

	"github.com/EricsAntony/salon/salon-shared/auth"
	"github.com/EricsAntony/salon/salon-shared/logger"
	"user-service/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	// Convert to shared config for compatibility
	sharedCfg := cfg.ToSharedConfig()

	// Init logger
	logger.Init(sharedCfg)
	log.Info().Str("service", cfg.Log.ServiceName).Msg("starting service")

	// Initialize tracing (optional - only if JAEGER_ENDPOINT is set)
	var tracingCleanup func()
	if jaegerEndpoint := os.Getenv("JAEGER_ENDPOINT"); jaegerEndpoint != "" {
		tracingCleanup = tracing.InitTracing(cfg.Log.ServiceName, jaegerEndpoint)
		defer tracingCleanup()
	}

	// Connect DB
	pool, err := pgxpool.New(context.Background(), cfg.DB.URL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer pool.Close()

	// Wire deps
	userRepo := repository.NewUserRepository(pool)
	otpRepo := repository.NewOTPRepository(pool)
	tokenRepo := repository.NewTokenRepository(pool)
	jwtMgr := auth.NewJWTManager(sharedCfg)
	userSvc := service.NewUserService(userRepo, otpRepo, tokenRepo, jwtMgr, cfg)
	h := api.NewHandler(userSvc, jwtMgr, sharedCfg)

	// Start OTP cleanup service
	cleanupSvc := cleanup.NewOTPCleanupService(otpRepo, 1*time.Hour, 24*time.Hour)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go cleanupSvc.Start(ctx)

	// Router
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(api.RequestLogger)
	r.Use(middleware.Recoverer)

	// Add Prometheus metrics endpoint
	r.Handle("/metrics", promhttp.Handler())

	h.RegisterRoutes(r)

	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info().Str("addr", srv.Addr).Msg("http server listening")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("http server error")
		}
	}()

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("graceful shutdown failed")
	}
	log.Info().Msg("server stopped")
}
