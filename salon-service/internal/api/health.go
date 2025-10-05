package api

import (
	"encoding/json"
	"net/http"
	"time"
)

type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp string            `json:"timestamp"`
	Service   string            `json:"service"`
	Version   string            `json:"version"`
	Checks    map[string]string `json:"checks"`
}

func (h *Handler) healthCheck(w http.ResponseWriter, r *http.Request) {
	// Check database connectivity
	dbStatus := "healthy"
	if err := h.svc.HealthCheck(r.Context()); err != nil {
		dbStatus = "unhealthy: " + err.Error()
	}

	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Service:   "salon-service",
		Version:   "1.0.0",
		Checks: map[string]string{
			"database": dbStatus,
		},
	}

	// Set overall status based on checks
	if dbStatus != "healthy" {
		response.Status = "unhealthy"
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
