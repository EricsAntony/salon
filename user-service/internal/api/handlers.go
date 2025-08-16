package api

import (
	"encoding/json"
	"encoding/hex"
	"crypto/rand"
	"net/http"
	"strings"
	"time"
	"net/mail"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"github.com/rs/zerolog/log"
	"github.com/EricsAntony/salon/salon-shared/auth"
	"github.com/EricsAntony/salon/salon-shared/config"
	"user-service/internal/service"
	"user-service/internal/errors"
	"github.com/EricsAntony/salon/salon-shared/utils"
)

type Handler struct {
	svc    service.UserService
	jwt    *auth.JWTManager
	cfg    *config.Config
}

func NewHandler(svc service.UserService, jwt *auth.JWTManager, cfg *config.Config) *Handler {
	return &Handler{svc: svc, jwt: jwt, cfg: cfg}
}

func (h *Handler) RegisterRoutes(r *chi.Mux) {
	r.Group(func(r chi.Router) {
		// OTP request with rate limit per IP
		r.With(httprate.LimitByIP(h.cfg.RateLimit.OTPRequestsPerMinute, time.Minute)).Post("/otp/request", h.requestOTP)
	})

	r.Post("/user/register", h.register)
	r.Post("/user/authenticate", h.authenticate)
	r.Post("/auth/refresh", h.refresh)

	// Health endpoints (no auth required)
	r.Get("/health", h.health)
	r.Get("/ready", h.readiness)

	r.Group(func(r chi.Router) {
		// Use shared JWT middleware for protected routes
		r.Use(h.jwt.Middleware())
		r.Get("/user/{id}", h.getUser)
		r.Post("/auth/revoke", h.revoke)
		r.Put("/users/{id}", h.updateUser)
		r.Delete("/users/{id}", h.deleteUser)
	})
}

type reqOTP struct { PhoneNumber string `json:"phone_number"` }

func (h *Handler) requestOTP(w http.ResponseWriter, r *http.Request) {
	var req reqOTP
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.PhoneNumber) == "" {
		writeErr(w, http.StatusBadRequest, "invalid request")
		return
	}
	if !utils.ValidPhone(req.PhoneNumber) {
		writeErr(w, http.StatusBadRequest, "invalid phone number format")
		return
	}
	code, err := h.svc.RequestOTP(r.Context(), req.PhoneNumber)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "otp_sent", "code": code})
}

type registerReq struct {
	PhoneNumber string  `json:"phone_number"`
	Name        string  `json:"name"`
	Gender      string  `json:"gender"`
	Email       *string `json:"email"`
	Location    *string `json:"location"`
	OTP         string  `json:"otp"`
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var req registerReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	if strings.TrimSpace(req.PhoneNumber) == "" || strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.Gender) == "" || strings.TrimSpace(req.OTP) == "" {
		writeErr(w, http.StatusBadRequest, "missing required fields")
		return
	}
	if !utils.ValidPhone(req.PhoneNumber) {
		writeErr(w, http.StatusBadRequest, "invalid phone number format")
		return
	}
	u, access, refresh, err := h.svc.Register(r.Context(), req.PhoneNumber, req.Name, req.Gender, req.Email, req.Location, req.OTP)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	h.setRefreshCookie(w, refresh)
	h.setCSRFCookie(w)
	writeJSON(w, http.StatusCreated, map[string]interface{}{"user": u, "access_token": access})
}

type authReq struct { PhoneNumber string `json:"phone_number"`; OTP string `json:"otp"` }

func (h *Handler) authenticate(w http.ResponseWriter, r *http.Request) {
	var req authReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	if strings.TrimSpace(req.PhoneNumber) == "" || strings.TrimSpace(req.OTP) == "" {
		writeErr(w, http.StatusBadRequest, "missing phone or otp")
		return
	}
	if !utils.ValidPhone(req.PhoneNumber) {
		writeErr(w, http.StatusBadRequest, "invalid phone number format")
		return
	}
	access, refresh, err := h.svc.Authenticate(r.Context(), req.PhoneNumber, req.OTP)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, err.Error())
		return
	}
	h.setRefreshCookie(w, refresh)
	h.setCSRFCookie(w)
	writeJSON(w, http.StatusOK, map[string]interface{}{"access_token": access})
}

func (h *Handler) getUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	u, err := h.svc.GetUser(r.Context(), id)
	if err != nil {
		writeErr(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, u)
}

func (h *Handler) refresh(w http.ResponseWriter, r *http.Request) {
	csrfHdr := r.Header.Get("X-CSRF-Token")
	csrfC, err := r.Cookie("csrf_token")
	if err != nil || csrfHdr == "" || csrfHdr != csrfC.Value {
		writeErr(w, http.StatusForbidden, "csrf validation failed")
		return
	}
	rc, err := r.Cookie("refresh_token")
	if err != nil || strings.TrimSpace(rc.Value) == "" {
		writeErr(w, http.StatusUnauthorized, "missing refresh token")
		return
	}
	access, newRefresh, err := h.svc.Refresh(r.Context(), rc.Value)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, err.Error())
		return
	}
	h.setRefreshCookie(w, newRefresh)
	h.setCSRFCookie(w)
	writeJSON(w, http.StatusOK, map[string]string{"access_token": access})
}

func (h *Handler) revoke(w http.ResponseWriter, r *http.Request) {
	uid, _ := r.Context().Value(auth.CtxUserID).(string)
	if strings.TrimSpace(uid) == "" {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if err := h.svc.Revoke(r.Context(), uid); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.clearRefreshCookie(w)
	h.clearCSRFCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

type updateReq struct {
	Name     *string `json:"name"`
	Gender   *string `json:"gender"`
	Email    *string `json:"email"`
	Location *string `json:"location"`
}

func (h *Handler) updateUser(w http.ResponseWriter, r *http.Request) {
	targetID := chi.URLParam(r, "id")
	uid, _ := r.Context().Value(auth.CtxUserID).(string)
	if uid == "" || uid != targetID {
		writeErr(w, http.StatusForbidden, "forbidden")
		return
	}
	var req updateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	if req.Email != nil {
		if _, err := mail.ParseAddress(strings.TrimSpace(*req.Email)); err != nil { writeErr(w, http.StatusBadRequest, "invalid email"); return }
	}
	if req.Gender != nil {
		g := strings.ToLower(strings.TrimSpace(*req.Gender))
		if g != "male" && g != "female" && g != "other" { writeErr(w, http.StatusBadRequest, "invalid gender"); return }
	}
	u, err := h.svc.UpdateUser(r.Context(), uid, req.Name, req.Gender, req.Email, req.Location)
	if err != nil { writeErr(w, http.StatusBadRequest, err.Error()); return }
	writeJSON(w, http.StatusOK, u)
}

func (h *Handler) deleteUser(w http.ResponseWriter, r *http.Request) {
	targetID := chi.URLParam(r, "id")
	uid, _ := r.Context().Value(auth.CtxUserID).(string)
	if uid == "" || uid != targetID { writeErr(w, http.StatusForbidden, "forbidden"); return }
	if err := h.svc.DeleteUser(r.Context(), uid, targetID); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
		"service": "user-service",
	})
}

func (h *Handler) readiness(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Check database connectivity
	if err := h.svc.HealthCheck(ctx); err != nil {
		log.Error().Err(err).Msg("readiness check failed")
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"status": "not ready",
			"error": "database connectivity failed",
		})
		return
	}
	
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ready",
		"service": "user-service",
	})
}

func (h *Handler) refreshCookieExpiry() time.Time {
	days := h.cfg.JWT.RefreshTTLDays
	if days <= 0 { days = 7 }
	return time.Now().Add(time.Duration(days) * 24 * time.Hour)
}

func (h *Handler) secureCookies() bool {
	env := strings.ToLower(strings.TrimSpace(h.cfg.Env))
	return env != "dev" && env != "docker" && env != "local"
}

func (h *Handler) setRefreshCookie(w http.ResponseWriter, token string) {
	cookie := &http.Cookie{
		Name:     "refresh_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   h.secureCookies(),
		SameSite: http.SameSiteStrictMode,
		Expires:  h.refreshCookieExpiry(),
	}
	http.SetCookie(w, cookie)
}

func (h *Handler) clearRefreshCookie(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   h.secureCookies(),
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Unix(0,0),
		MaxAge:   -1,
	}
	http.SetCookie(w, cookie)
}

func (h *Handler) setCSRFCookie(w http.ResponseWriter) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil { log.Warn().Err(err).Msg("csrf token gen failed") }
	v := hex.EncodeToString(b)
	cookie := &http.Cookie{
		Name:     "csrf_token",
		Value:    v,
		Path:     "/",
		HttpOnly: false,
		Secure:   h.secureCookies(),
		SameSite: http.SameSiteLaxMode,
		Expires:  h.refreshCookieExpiry(),
	}
	http.SetCookie(w, cookie)
}

func (h *Handler) clearCSRFCookie(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     "csrf_token",
		Value:    "",
		Path:     "/",
		HttpOnly: false,
		Secure:   h.secureCookies(),
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Unix(0,0),
		MaxAge:   -1,
	}
	http.SetCookie(w, cookie)
}

func writeJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}

func writeAPIError(w http.ResponseWriter, err error) {
	apiErr := errors.MapToAPIError(err)
	log.Error().Err(err).Int("status_code", apiErr.Code).Str("error_type", apiErr.Type).Msg("api error")
	writeJSON(w, apiErr.Code, apiErr)
}

// RequestLogger adds request_id to logs and basic structured request logs
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := middleware.GetReqID(r.Context())
		log.Info().Str("method", r.Method).Str("path", r.URL.Path).Str("request_id", reqID).Msg("incoming request")
		next.ServeHTTP(w, r)
	})
}
