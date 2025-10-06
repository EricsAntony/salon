package api

import (
	"net/http"

	"payment-service/internal/service"

	"github.com/rs/zerolog/log"
	"salon-shared/errors"
	"salon-shared/utils"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	paymentService service.PaymentService
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(paymentService service.PaymentService) *HealthHandler {
	return &HealthHandler{
		paymentService: paymentService,
	}
}

// Health handles GET /api/v1/health
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	response, err := h.paymentService.HealthCheck(r.Context())
	if err != nil {
		log.Error().Err(err).Msg("Health check failed")
		errors.WriteAPIError(w, errors.MapToAPIError(err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, response)
}

// Ready handles GET /api/v1/ready
func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	// Ready check is similar to health but can be more strict
	response, err := h.paymentService.HealthCheck(r.Context())
	if err != nil {
		log.Error().Err(err).Msg("Readiness check failed")
		w.WriteHeader(http.StatusServiceUnavailable)
		utils.WriteJSON(w, http.StatusServiceUnavailable, map[string]string{
			"status": "not ready",
			"error":  err.Error(),
		})
		return
	}

	// Check if database is healthy for readiness
	if response.Database != "healthy" {
		w.WriteHeader(http.StatusServiceUnavailable)
		utils.WriteJSON(w, http.StatusServiceUnavailable, map[string]string{
			"status":   "not ready",
			"database": response.Database,
		})
		return
	}

	utils.WriteJSON(w, http.StatusOK, response)
}
