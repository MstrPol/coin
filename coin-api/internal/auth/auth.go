package auth

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"strings"

	"coin.local/coin-api/internal/config"
)

func Bearer(cfg config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfg.AuthDisabled {
				next.ServeHTTP(w, r)
				return
			}
			if cfg.APIToken == "" {
				writeJSON(w, http.StatusInternalServerError, map[string]string{
					"error": "auth not configured",
				})
				return
			}
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				writeJSON(w, http.StatusUnauthorized, map[string]string{
					"error": "missing bearer token",
				})
				return
			}
			token := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
			if subtle.ConstantTimeCompare([]byte(token), []byte(cfg.APIToken)) != 1 {
				writeJSON(w, http.StatusUnauthorized, map[string]string{
					"error": "invalid token",
				})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
