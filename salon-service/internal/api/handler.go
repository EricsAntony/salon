package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	sharedAuth "github.com/EricsAntony/salon/salon-shared/auth"
	sharedConfig "github.com/EricsAntony/salon/salon-shared/config"
	"github.com/EricsAntony/salon/salon-shared/logger"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"
	"salon-service/internal/model"
	"salon-service/internal/repository"
	"salon-service/internal/service"
)

type Handler struct {
	svc service.SalonService
	jwt *sharedAuth.JWTManager
}

func NewHandler(cfg *sharedConfig.Config, svc service.SalonService) *Handler {
	logger.Init(cfg)
	return &Handler{
		svc: svc,
		jwt: sharedAuth.NewJWTManager(cfg),
	}
}

func (h *Handler) Routes() *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)

	r.Get("/health", h.health)
	r.Get("/ready", h.ready)

	r.Group(func(r chi.Router) {
		r.Post("/staff/otp", h.requestStaffOTP)
		r.Post("/staff/auth", h.authenticateStaff)
		r.Post("/staff/refresh", h.refreshStaffSession)
	})

	r.Group(func(r chi.Router) {
		r.Use(h.jwt.Middleware())
		r.Route("/salons", func(r chi.Router) {
			r.Post("/", h.createSalon)
			r.Get("/", h.listSalons)
			r.Route("/{salonID}", func(r chi.Router) {
				r.Get("/", h.getSalon)
				r.Put("/", h.updateSalon)
				r.Delete("/", h.deleteSalon)

				r.Route("/branches", func(r chi.Router) {
					r.Post("/", h.createBranch)
					r.Get("/", h.listBranches)
					r.Route("/{branchID}", func(r chi.Router) {
						r.Get("/", h.getBranch)
						r.Put("/", h.updateBranch)
						r.Delete("/", h.deleteBranch)
					})
				})

				r.Route("/categories", func(r chi.Router) {
					r.Post("/", h.createCategory)
					r.Get("/", h.listCategories)
					r.Route("/{categoryID}", func(r chi.Router) {
						r.Put("/", h.updateCategory)
						r.Delete("/", h.deleteCategory)
					})
				})

				r.Route("/services", func(r chi.Router) {
					r.Post("/", h.createService)
					r.Get("/", h.listServices)
					r.Route("/{serviceID}", func(r chi.Router) {
						r.Put("/", h.updateService)
						r.Delete("/", h.deleteService)
					})
				})

				r.Route("/staff", func(r chi.Router) {
					r.Post("/", h.createStaff)
					r.Get("/", h.listStaff)
					r.Route("/{staffID}", func(r chi.Router) {
						r.Put("/", h.updateStaff)
						r.Delete("/", h.deleteStaff)
						r.Post("/services", h.setStaffServices)
						r.Get("/services", h.listStaffServices)
					})
				})
			})
		})
	})

	return r
}

func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"service": "salon-service",
	})
}

func (h *Handler) ready(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.HealthCheck(r.Context()); err != nil {
		log.Error().Err(err).Msg("readiness check failed")
		writeError(w, http.StatusServiceUnavailable, "database not ready")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "ready",
		"service": "salon-service",
	})
}

func (h *Handler) createSalon(w http.ResponseWriter, r *http.Request) {
	var req createSalonRequest
	if err := decodeRequest(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	params := createSalonRequest(req).toCreateParams()
	salon, err := h.svc.CreateSalon(r.Context(), params)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, salon)
}

func (h *Handler) getSalon(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(chi.URLParam(r, "salonID"))
	salon, err := h.svc.GetSalon(r.Context(), id)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, salon)
}

func (h *Handler) listSalons(w http.ResponseWriter, r *http.Request) {
	salons, err := h.svc.ListSalons(r.Context())
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, salons)
}

func (h *Handler) updateSalon(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(chi.URLParam(r, "salonID"))
	var req updateSalonRequest
	if err := decodeRequest(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	params := updateSalonRequest(req).toUpdateParams(id)
	salon, err := h.svc.UpdateSalon(r.Context(), params)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, salon)
}

func (h *Handler) deleteSalon(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(chi.URLParam(r, "salonID"))
	if err := h.svc.DeleteSalon(r.Context(), id); err != nil {
		handleServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) createBranch(w http.ResponseWriter, r *http.Request) {
	salonID := strings.TrimSpace(chi.URLParam(r, "salonID"))
	var req createBranchRequest
	if err := decodeRequest(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	params := createBranchRequest(req).toCreateParams(salonID)
	branch, err := h.svc.CreateBranch(r.Context(), params)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, branch)
}

func (h *Handler) getBranch(w http.ResponseWriter, r *http.Request) {
	salonID := strings.TrimSpace(chi.URLParam(r, "salonID"))
	branchID := strings.TrimSpace(chi.URLParam(r, "branchID"))
	branch, err := h.svc.GetBranch(r.Context(), salonID, branchID)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, branch)
}

func (h *Handler) listBranches(w http.ResponseWriter, r *http.Request) {
	salonID := strings.TrimSpace(chi.URLParam(r, "salonID"))
	branches, err := h.svc.ListBranches(r.Context(), salonID)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, branches)
}

func (h *Handler) updateBranch(w http.ResponseWriter, r *http.Request) {
	salonID := strings.TrimSpace(chi.URLParam(r, "salonID"))
	branchID := strings.TrimSpace(chi.URLParam(r, "branchID"))
	var req updateBranchRequest
	if err := decodeRequest(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	params := updateBranchRequest(req).toUpdateParams(salonID, branchID)
	branch, err := h.svc.UpdateBranch(r.Context(), params)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, branch)
}

func (h *Handler) deleteBranch(w http.ResponseWriter, r *http.Request) {
	salonID := strings.TrimSpace(chi.URLParam(r, "salonID"))
	branchID := strings.TrimSpace(chi.URLParam(r, "branchID"))
	if err := h.svc.DeleteBranch(r.Context(), salonID, branchID); err != nil {
		handleServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) createCategory(w http.ResponseWriter, r *http.Request) {
	salonID := strings.TrimSpace(chi.URLParam(r, "salonID"))
	var req createCategoryRequest
	if err := decodeRequest(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	params := req.toCreateParams(salonID)
	category, err := h.svc.CreateCategory(r.Context(), params)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, category)
}

func (h *Handler) listCategories(w http.ResponseWriter, r *http.Request) {
	salonID := strings.TrimSpace(chi.URLParam(r, "salonID"))
	categories, err := h.svc.ListCategories(r.Context(), salonID)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, categories)
}

func (h *Handler) updateCategory(w http.ResponseWriter, r *http.Request) {
	salonID := strings.TrimSpace(chi.URLParam(r, "salonID"))
	categoryID := strings.TrimSpace(chi.URLParam(r, "categoryID"))
	var req updateCategoryRequest
	if err := decodeRequest(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	params := req.toUpdateParams(salonID, categoryID)
	category, err := h.svc.UpdateCategory(r.Context(), params)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, category)
}

func (h *Handler) deleteCategory(w http.ResponseWriter, r *http.Request) {
	salonID := strings.TrimSpace(chi.URLParam(r, "salonID"))
	categoryID := strings.TrimSpace(chi.URLParam(r, "categoryID"))
	if err := h.svc.DeleteCategory(r.Context(), salonID, categoryID); err != nil {
		handleServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) createService(w http.ResponseWriter, r *http.Request) {
	salonID := strings.TrimSpace(chi.URLParam(r, "salonID"))
	var req createServiceRequest
	if err := decodeRequest(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	params := req.toCreateParams(salonID)
	service, err := h.svc.CreateService(r.Context(), params)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, service)
}

func (h *Handler) listServices(w http.ResponseWriter, r *http.Request) {
	salonID := strings.TrimSpace(chi.URLParam(r, "salonID"))
	var categoryID *string
	if val := strings.TrimSpace(r.URL.Query().Get("category_id")); val != "" {
		categoryID = &val
	}
	services, err := h.svc.ListServices(r.Context(), salonID, categoryID)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, services)
}

func (h *Handler) updateService(w http.ResponseWriter, r *http.Request) {
	salonID := strings.TrimSpace(chi.URLParam(r, "salonID"))
	serviceID := strings.TrimSpace(chi.URLParam(r, "serviceID"))
	var req updateServiceRequest
	if err := decodeRequest(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	params := req.toUpdateParams(salonID, serviceID)
	service, err := h.svc.UpdateService(r.Context(), params)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, service)
}

func (h *Handler) deleteService(w http.ResponseWriter, r *http.Request) {
	salonID := strings.TrimSpace(chi.URLParam(r, "salonID"))
	serviceID := strings.TrimSpace(chi.URLParam(r, "serviceID"))
	if err := h.svc.DeleteService(r.Context(), salonID, serviceID); err != nil {
		handleServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) createStaff(w http.ResponseWriter, r *http.Request) {
	salonID := strings.TrimSpace(chi.URLParam(r, "salonID"))
	var req createStaffRequest
	if err := decodeRequest(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	params := req.toCreateParams(salonID)
	staff, err := h.svc.CreateStaff(r.Context(), params)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, staff)
}

func (h *Handler) listStaff(w http.ResponseWriter, r *http.Request) {
	salonID := strings.TrimSpace(chi.URLParam(r, "salonID"))
	var status *model.StaffStatus
	if val := strings.TrimSpace(r.URL.Query().Get("status")); val != "" {
		st := model.StaffStatus(val)
		status = &st
	}
	staff, err := h.svc.ListStaff(r.Context(), salonID, status)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, staff)
}

func (h *Handler) updateStaff(w http.ResponseWriter, r *http.Request) {
	salonID := strings.TrimSpace(chi.URLParam(r, "salonID"))
	staffID := strings.TrimSpace(chi.URLParam(r, "staffID"))
	var req updateStaffRequest
	if err := decodeRequest(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	params := req.toUpdateParams(salonID, staffID)
	staff, err := h.svc.UpdateStaff(r.Context(), params)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, staff)
}

func (h *Handler) deleteStaff(w http.ResponseWriter, r *http.Request) {
	salonID := strings.TrimSpace(chi.URLParam(r, "salonID"))
	staffID := strings.TrimSpace(chi.URLParam(r, "staffID"))
	if err := h.svc.DeleteStaff(r.Context(), salonID, staffID); err != nil {
		handleServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) setStaffServices(w http.ResponseWriter, r *http.Request) {
	salonID := strings.TrimSpace(chi.URLParam(r, "salonID"))
	staffID := strings.TrimSpace(chi.URLParam(r, "staffID"))
	var req setStaffServicesRequest
	if err := decodeRequest(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.svc.SetStaffServices(r.Context(), salonID, staffID, req.ServiceIDs); err != nil {
		handleServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) listStaffServices(w http.ResponseWriter, r *http.Request) {
	staffID := strings.TrimSpace(chi.URLParam(r, "staffID"))
	ids, err := h.svc.ListStaffServices(r.Context(), staffID)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"service_ids": ids})
}

func (h *Handler) requestStaffOTP(w http.ResponseWriter, r *http.Request) {
	var req requestStaffOTPRequest
	if err := decodeRequest(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.svc.RequestStaffOTP(r.Context(), service.RequestStaffOTPParams{PhoneNumber: req.PhoneNumber}); err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]any{"success": true})
}

func (h *Handler) authenticateStaff(w http.ResponseWriter, r *http.Request) {
	var req authenticateStaffRequest
	if err := decodeRequest(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	res, err := h.svc.AuthenticateStaff(r.Context(), service.AuthenticateStaffParams{PhoneNumber: req.PhoneNumber, OTP: req.OTP})
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (h *Handler) refreshStaffSession(w http.ResponseWriter, r *http.Request) {
	var req refreshStaffSessionRequest
	if err := decodeRequest(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	claims, err := h.jwt.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid token")
		return
	}
	res, err := h.svc.RefreshStaffSession(r.Context(), claims.UserID, req.RefreshToken)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func decodeRequest(r *http.Request, v any) error {
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return err
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]any{
		"success": false,
		"error": map[string]any{
			"code":    status,
			"type":    http.StatusText(status),
			"message": message,
		},
	})
}

func handleServiceError(w http.ResponseWriter, err error) {
	switch e := err.(type) {
	case service.ValidationErrors:
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"error": map[string]any{
				"code":    http.StatusBadRequest,
				"type":    "validation_error",
				"message": "validation failed",
				"details": e,
			},
		})
	case service.ConflictError:
		writeError(w, http.StatusConflict, e.Error())
	default:
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
	}
}
