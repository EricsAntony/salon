package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/EricsAntony/salon/salon-shared/auth"
	"github.com/rs/zerolog/log"
)

// UserTypeMiddleware validates JWT tokens for specific user types
func UserTypeMiddleware(jwt *auth.JWTManager, expectedUserType string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authz := r.Header.Get("Authorization")
			parts := strings.SplitN(authz, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				http.Error(w, "missing bearer token", http.StatusUnauthorized)
				return
			}

			claims, err := jwt.ValidateAccessToken(parts[1])
			if err != nil {
				log.Warn().Err(err).Msg("invalid access token")
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			// Validate user type
			if claims.UserType != expectedUserType {
				log.Warn().
					Str("expected_type", expectedUserType).
					Str("actual_type", claims.UserType).
					Msg("unauthorized user type")
				http.Error(w, "unauthorized user type", http.StatusForbidden)
				return
			}

			ctx := context.WithValue(r.Context(), auth.CtxUserID, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// SalonUserMiddleware validates that the JWT token is of type "salon_USER"
func SalonUserMiddleware(jwt *auth.JWTManager) func(http.Handler) http.Handler {
	return UserTypeMiddleware(jwt, auth.UserTypeSalon)
}

// CustomerMiddleware validates that the JWT token is of type "customer"
func CustomerMiddleware(jwt *auth.JWTManager) func(http.Handler) http.Handler {
	return UserTypeMiddleware(jwt, auth.UserTypeCustomer)
}

// RequireAuthMiddleware ensures a valid JWT token is present (any type)
func RequireAuthMiddleware(jwt *auth.JWTManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authz := r.Header.Get("Authorization")
			parts := strings.SplitN(authz, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				http.Error(w, "missing bearer token", http.StatusUnauthorized)
				return
			}

			claims, err := jwt.ValidateAccessToken(parts[1])
			if err != nil {
				log.Warn().Err(err).Msg("invalid access token")
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), auth.CtxUserID, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
