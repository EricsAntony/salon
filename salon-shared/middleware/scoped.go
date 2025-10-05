package middleware

import (
	"context"
	"net/http"

	"github.com/EricsAntony/salon/salon-shared/auth"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

// SalonRepository interface for checking staff salon access
type SalonRepository interface {
	StaffHasAccessToSalon(ctx context.Context, staffID, salonID string) (bool, error)
}

// SalonScopedMiddleware ensures staff can only access their own salon's data
func SalonScopedMiddleware(repo SalonRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := r.Context().Value(auth.CtxUserID).(string)
			salonID := chi.URLParam(r, "salonID")
			
			if salonID == "" {
				http.Error(w, "salon ID required", http.StatusBadRequest)
				return
			}

			// Check if the staff belongs to this salon
			hasAccess, err := repo.StaffHasAccessToSalon(r.Context(), userID, salonID)
			if err != nil {
				log.Error().Err(err).Str("staff_id", userID).Str("salon_id", salonID).Msg("failed to check salon access")
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}

			if !hasAccess {
				log.Warn().Str("staff_id", userID).Str("salon_id", salonID).Msg("unauthorized salon access attempt")
				http.Error(w, "access denied to salon", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// UserScopedMiddleware ensures users can only access their own data
func UserScopedMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := r.Context().Value(auth.CtxUserID).(string)
			targetUserID := chi.URLParam(r, "id")
			
			if targetUserID == "" {
				http.Error(w, "user ID required", http.StatusBadRequest)
				return
			}

			// Check if the user is trying to access their own data
			if userID != targetUserID {
				log.Warn().Str("user_id", userID).Str("target_user_id", targetUserID).Msg("unauthorized user access attempt")
				http.Error(w, "access denied: can only access own data", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
