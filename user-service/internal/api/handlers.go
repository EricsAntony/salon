package api

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"user-service/internal/api/interfaces"
	appErrors "user-service/internal/errors"
	"user-service/internal/service"

	"github.com/EricsAntony/salon/salon-shared/auth"
	"github.com/EricsAntony/salon/salon-shared/config"
	sharedMiddleware "github.com/EricsAntony/salon/salon-shared/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"
)

type Handler struct {
	svc         service.UserService
	jwt         *auth.JWTManager
	cfg         *config.Config
	rateLimiter *sharedMiddleware.RateLimiter
}

func NewHandler(svc service.UserService, jwt *auth.JWTManager, cfg *config.Config) *Handler {
	// Create rate limiter: 3 OTP requests per 10 minutes per IP
	rateLimiter := sharedMiddleware.NewRateLimiter(3, 10*time.Minute)
	
	return &Handler{
		svc:         svc,
		jwt:         jwt,
		cfg:         cfg,
		rateLimiter: rateLimiter,
	}
}

func (h *Handler) RegisterRoutes(r *chi.Mux) {
	// Add shared audit middleware with user-service configuration
	auditConfig := sharedMiddleware.DefaultAuditConfig("user-service")
	r.Use(sharedMiddleware.AuditMiddleware(auditConfig))
	
	r.Group(func(r chi.Router) {
		// OTP request with enhanced rate limiting
		r.With(sharedMiddleware.OTPRateLimitMiddleware(h.rateLimiter)).Post("/otp/request", h.requestOTP)
	})

	r.Post("/user/register", h.register)
	r.Post("/user/authenticate", h.authenticate)
	r.Post("/auth/refresh", h.refresh)

	// Health endpoints (no auth required)
	r.Get("/health", h.health)
	r.Get("/ready", h.readiness)

	r.Group(func(r chi.Router) {
		// Use customer-specific middleware for protected routes
		r.Use(sharedMiddleware.CustomerMiddleware(h.jwt))
		// Enforce CSRF on unsafe methods for protected routes
		r.Use(RequireCSRF)
		
		// User-scoped routes - users can only access their own data
		r.With(sharedMiddleware.UserScopedMiddleware()).Get("/user/{id}", h.getUser)
		r.With(sharedMiddleware.UserScopedMiddleware()).Put("/users/{id}", h.updateUser)
		r.With(sharedMiddleware.UserScopedMiddleware()).Delete("/users/{id}", h.deleteUser)
		
		// Auth operations (no user scoping needed)
		r.Post("/auth/revoke", h.revoke)
	})
}

func (h *Handler) requestOTP(w http.ResponseWriter, r *http.Request) {
	var req interfaces.ReqOTP
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid request")
		return
	}
	if err := req.ValidateStrict(); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	code, err := h.svc.RequestOTP(r.Context(), req.PhoneNumber)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "otp_sent", "code": code})
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var req interfaces.RegisterReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	if err := req.ValidateStrict(); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	params := service.RegisterParams{
		Phone:    req.PhoneNumber,
		Name:     req.Name,
		Gender:   req.Gender,
		Email:    req.Email,
		Location: req.Location,
		Lat:      req.Lat,
		Lng:      req.Lng,
		OTP:      req.OTP,
	}
	u, access, refresh, err := h.svc.Register(r.Context(), params)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	h.setRefreshCookie(w, refresh)
	h.setCSRFCookie(w)
	writeJSON(w, http.StatusCreated, map[string]interface{}{"user_id": u.ID, "access_token": access})
}

func (h *Handler) authenticate(w http.ResponseWriter, r *http.Request) {
	var req interfaces.AuthReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	if err := req.ValidateStrict(); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	userId, access, refresh, err := h.svc.Authenticate(r.Context(), req.PhoneNumber, req.OTP)
	if err != nil {
		writeAPIError(w, err)
		return
	}
	h.setRefreshCookie(w, refresh)
	h.setCSRFCookie(w)
	writeJSON(w, http.StatusOK, map[string]interface{}{"user_id": userId, "access_token": access})
}

func (h *Handler) getUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(strings.TrimSpace(id)); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid user id")
		return
	}
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
	if err != nil || strings.TrimSpace(csrfHdr) == "" || strings.TrimSpace(csrfC.Value) == "" {
		writeErr(w, http.StatusForbidden, "csrf validation failed")
		return
	}
	// Validate CSRF token format (expect 64 hex chars)
	if len(csrfHdr) != 64 {
		writeErr(w, http.StatusForbidden, "csrf validation failed")
		return
	}
	if _, err := hex.DecodeString(csrfHdr); err != nil {
		writeErr(w, http.StatusForbidden, "csrf validation failed")
		return
	}
	if csrfHdr != csrfC.Value {
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

func (h *Handler) updateUser(w http.ResponseWriter, r *http.Request) {
	targetID := chi.URLParam(r, "id")
	if _, err := uuid.Parse(strings.TrimSpace(targetID)); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid user id")
		return
	}
	uid, _ := r.Context().Value(auth.CtxUserID).(string)
	if uid == "" {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req interfaces.UpdateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	if err := req.ValidateStrict(); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	up := service.UpdateUserParams{
		UserID:   uid,
		Location: req.Location,
		Lat:      req.Lat,
		Lng:      req.Lng,
	}
	u, err := h.svc.UpdateUser(r.Context(), up)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, u)
}

func (h *Handler) deleteUser(w http.ResponseWriter, r *http.Request) {
	targetID := chi.URLParam(r, "id")
	if _, err := uuid.Parse(strings.TrimSpace(targetID)); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid user id")
		return
	}
	uid, _ := r.Context().Value(auth.CtxUserID).(string)
	if uid == "" || uid != targetID {
		writeErr(w, http.StatusForbidden, "forbidden")
		return
	}
	if err := h.svc.DeleteUser(r.Context(), uid, targetID); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"service": "user-service",
	})
}

func (h *Handler) readiness(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Check database connectivity
	if err := h.svc.HealthCheck(ctx); err != nil {
		log.Error().Err(err).Msg("readiness check failed")
		writeErr(w, http.StatusServiceUnavailable, "database connectivity failed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "ready",
		"service": "user-service",
	})
}

func (h *Handler) refreshCookieExpiry() time.Time {
	days := h.cfg.JWT.RefreshTTLDays
	if days <= 0 {
		days = 7
	}
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
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
	}
	http.SetCookie(w, cookie)
}

func (h *Handler) setCSRFCookie(w http.ResponseWriter) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		log.Warn().Err(err).Msg("csrf token gen failed")
	}
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
		Expires:  time.Unix(0, 0),
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
	writeJSON(w, code, map[string]interface{}{
		"success": false,
		"error": map[string]interface{}{
			"code":    code,
			"message": msg,
		},
	})
}

func writeAPIError(w http.ResponseWriter, err error) {
	apiErr := appErrors.MapToAPIError(err)
	log.Error().Err(err).Int("status_code", apiErr.Code).Str("error_type", apiErr.Type).Msg("api error")
	writeJSON(w, apiErr.Code, map[string]interface{}{
		"success": false,
		"error": map[string]interface{}{
			"code":    apiErr.Code,
			"message": apiErr.Message,
			"type":    apiErr.Type,
		},
	})
}

// RequireCSRF validates CSRF token for unsafe HTTP methods by comparing
// the X-CSRF-Token header with the csrf_token cookie.
func RequireCSRF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
			csrfHdr := strings.TrimSpace(r.Header.Get("X-CSRF-Token"))
			csrfC, err := r.Cookie("csrf_token")
			if err != nil || csrfHdr == "" || strings.TrimSpace(csrfC.Value) == "" {
				writeErr(w, http.StatusForbidden, "csrf validation failed")
				return
			}
			// Validate format (expect 64 hex chars)
			if len(csrfHdr) != 64 {
				writeErr(w, http.StatusForbidden, "csrf validation failed")
				return
			}
			if _, err := hex.DecodeString(csrfHdr); err != nil {
				writeErr(w, http.StatusForbidden, "csrf validation failed")
				return
			}
			if subtle.ConstantTimeCompare([]byte(csrfHdr), []byte(csrfC.Value)) != 1 {
				writeErr(w, http.StatusForbidden, "csrf validation failed")
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// RequestLogger adds request_id to logs and basic structured request logs
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := middleware.GetReqID(r.Context())
		log.Info().Str("method", r.Method).Str("path", r.URL.Path).Str("request_id", reqID).Msg("incoming request")
		next.ServeHTTP(w, r)
	})
}