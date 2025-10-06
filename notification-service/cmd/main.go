package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"notification-service/internal/api"
	"notification-service/internal/config"
	"notification-service/internal/consumer"
	"notification-service/internal/db"
	"notification-service/internal/provider"
	"notification-service/internal/repository"
	"notification-service/internal/service"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Configure structured logging
	zerolog.TimeFieldFormat = time.RFC3339
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Set log level
	level, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		log.Warn().Str("level", cfg.LogLevel).Msg("Invalid log level, using info")
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	log.Info().
		Str("env", cfg.Environment).
		Str("port", cfg.Port).
		Msg("Starting notification service")

	// Initialize database
	database, err := db.NewConnection(cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer database.Close()

	// Run migrations
	if err := db.RunMigrations(database, cfg.MigrationsPath); err != nil {
		log.Fatal().Err(err).Msg("Failed to run database migrations")
	}

	// Initialize repositories
	notificationRepo := repository.NewNotificationRepository(database)

	// Initialize notification providers
	providerManager := provider.NewProviderManager(cfg)

	// Initialize services
	notificationService := service.NewNotificationService(notificationRepo, providerManager)

	// Initialize HTTP server
	router := api.SetupRoutes(notificationService)
	httpServer := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}
	
	// Initialize event consumers
	eventConsumer := consumer.NewEventConsumer(notificationService)

	// Start services
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())

	// Start HTTP server
	wg.Add(1)
	go func() {
		defer wg.Done()
		
		httpServer.ReadTimeout = 30 * time.Second
		httpServer.WriteTimeout = 30 * time.Second
		httpServer.IdleTimeout = 120 * time.Second

		log.Info().Str("addr", httpServer.Addr).Msg("Notification service HTTP server listening")
		
		go func() {
			<-ctx.Done()
			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer shutdownCancel()
			httpServer.Shutdown(shutdownCtx)
		}()

		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Failed to start HTTP server")
		}
	}()

	// Start event consumer if configured
	if eventConsumer != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			log.Info().Str("broker", cfg.MessageBroker).Msg("Starting event consumer")
			if err := eventConsumer.Start(ctx); err != nil {
				log.Error().Err(err).Msg("Event consumer failed")
			}
		}()
	}

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down notification service...")

	// Cancel context to stop all services
	cancel()

	// Wait for all services to stop
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	// Wait for graceful shutdown or timeout
	select {
	case <-done:
		log.Info().Msg("Notification service stopped gracefully")
	case <-time.After(45 * time.Second):
		log.Warn().Msg("Notification service shutdown timeout")
	}
}
