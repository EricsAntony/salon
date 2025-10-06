package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"booking-service/internal/api"
	"booking-service/internal/config"
	"booking-service/internal/db"
	"booking-service/internal/repository"
	"booking-service/internal/service"

	"github.com/EricsAntony/salon/salon-shared/auth"
	"github.com/EricsAntony/salon/salon-shared/middleware"
	sharedconfig "github.com/EricsAntony/salon/salon-shared/config"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Initialize configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Setup logger
	setupLogger(cfg.Environment)

	log.Info().Msg("Starting booking-service...")

	// Initialize database
	database, err := db.NewConnection(cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer database.Close()

	// Run migrations
	if err := db.RunMigrations(cfg.DatabaseURL); err != nil {
		log.Fatal().Err(err).Msg("Failed to run database migrations")
	}

	// Initialize JWT manager
	jwtConfig := &sharedconfig.Config{
		JWT: struct {
			AccessSecret     string
			RefreshSecret    string
			AccessTTLMinutes int
			RefreshTTLDays   int
		}{
			AccessSecret:     cfg.JWTAccessSecret,
			RefreshSecret:    cfg.JWTRefreshSecret,
			AccessTTLMinutes: 60,
			RefreshTTLDays:   7,
		},
	}
	jwtManager := auth.NewJWTManager(jwtConfig)

	// Initialize repositories
	bookingRepo := repository.NewBookingRepository(database)
	
	// Initialize services
	bookingService := service.NewBookingService(bookingRepo, cfg)
	
	// Initialize handlers
	handlers := api.NewHandlers(bookingService)

	// Setup router
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Timeout(60 * time.Second))

	// Audit middleware
	auditConfig := middleware.DefaultAuditConfig("booking-service")
	r.Use(middleware.AuditMiddleware(auditConfig))

	// Health endpoints
	r.Get("/health", handlers.Health)
	r.Get("/ready", handlers.Ready)

	// API routes with authentication
	r.Route("/api/v1", func(r chi.Router) {
		// Customer authentication middleware
		r.Use(middleware.CustomerMiddleware(jwtManager))
		
		// Booking routes
		r.Post("/bookings/initiate", handlers.InitiateBooking)
		r.Get("/stylists/{stylistId}/availability", handlers.GetStylistAvailability)
		r.Post("/bookings/summary", handlers.CalculateBookingSummary)
		r.Post("/bookings/confirm", handlers.ConfirmBooking)
		r.Get("/bookings/{bookingId}", handlers.GetBooking)
		r.Get("/bookings/user/{userId}", handlers.GetUserBookings)
		r.Patch("/bookings/{bookingId}/cancel", handlers.CancelBooking)
		r.Patch("/bookings/{bookingId}/reschedule", handlers.RescheduleBooking)
		
		// Branch configuration
		r.Get("/branches/{branchId}/config", handlers.GetBranchConfig)
	})

	// Start server
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: r,
	}

	// Graceful shutdown
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Server failed to start")
		}
	}()

	log.Info().Int("port", cfg.Port).Msg("Booking service started successfully")

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server exited")
}

func setupLogger(env string) {
	if env == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
}
