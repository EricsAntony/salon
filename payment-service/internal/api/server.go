package api

import (
	"time"

	"payment-service/internal/config"
	"payment-service/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

// NewServer creates a new HTTP server with all routes configured
func NewServer(paymentService service.PaymentService, cfg *config.Config) *chi.Mux {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Timeout(30 * time.Second))

	// CORS configuration
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"}, // Configure based on your needs
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Idempotency-Key"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Initialize handlers
	paymentHandler := NewPaymentHandler(paymentService)
	webhookHandler := NewWebhookHandler(paymentService)
	healthHandler := NewHealthHandler(paymentService)

	// Routes
	r.Route("/api/v1", func(r chi.Router) {
		// Health endpoints
		r.Get("/health", healthHandler.Health)
		r.Get("/ready", healthHandler.Ready)

		// Payment endpoints
		r.Route("/payments", func(r chi.Router) {
			r.Post("/initiate", paymentHandler.InitiatePayment)
			r.Post("/confirm", paymentHandler.ConfirmPayment)
			r.Get("/{paymentID}", paymentHandler.GetPayment)
			r.Post("/{paymentID}/retry", paymentHandler.RetryPayment)
			
			// Refund endpoints
			r.Post("/{paymentID}/refund", paymentHandler.RefundPayment)
			r.Get("/{paymentID}/refunds", paymentHandler.GetRefunds)
		})

		// Booking-specific endpoints
		r.Route("/bookings", func(r chi.Router) {
			r.Get("/{bookingID}/payments", paymentHandler.GetPaymentsByBooking)
		})

		// User-specific endpoints
		r.Route("/users", func(r chi.Router) {
			r.Get("/{userID}/payments", paymentHandler.GetPaymentsByUser)
		})

		// Refund endpoints
		r.Route("/refunds", func(r chi.Router) {
			r.Get("/{refundID}", paymentHandler.GetRefund)
		})
	})

	// Webhook endpoints (separate from API versioning)
	r.Route("/webhooks", func(r chi.Router) {
		r.Post("/stripe", webhookHandler.StripeWebhook)
		r.Post("/razorpay", webhookHandler.RazorpayWebhook)
	})

	return r
}
