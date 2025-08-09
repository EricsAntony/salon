package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"
)

type ctxKey string

const ctxUserID ctxKey = "uid"

func (h *Handler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authz := r.Header.Get("Authorization")
		parts := strings.SplitN(authz, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			writeErr(w, http.StatusUnauthorized, "missing bearer token")
			return
		}
		claims, err := h.jwt.ValidateAccessToken(parts[1])
		if err != nil {
			log.Warn().Err(err).Msg("invalid access token")
			writeErr(w, http.StatusUnauthorized, "invalid token")
			return
		}
		// add user id to context for downstream handlers
		ctx := context.WithValue(r.Context(), ctxUserID, claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
