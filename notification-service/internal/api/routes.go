package api

import (
	"notification-service/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func SetupRoutes(notificationService *service.NotificationService) *chi.Mux {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	// Create handlers
	notificationHandler := NewNotificationHandler(notificationService)

	// Health check
	r.Get("/health", notificationHandler.Health)

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/notifications", func(r chi.Router) {
			r.Post("/send", notificationHandler.SendNotification)
			r.Get("/", notificationHandler.GetNotifications)
			r.Get("/{id}", notificationHandler.GetNotification)
		})
	})

	return r
}
