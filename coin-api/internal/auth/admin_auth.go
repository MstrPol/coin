package auth

import (
	"context"
	"crypto/subtle"
	"net/http"
	"strings"

	"coin.local/coin-api/internal/config"
)

const adminKeyHeader = "X-API-Key"

var localDevPrincipal = Principal{
	Subject:    "local-dev",
	Roles:      []Role{RoleAdmin},
	AuthMethod: "local",
}

// AdminAuth authenticates admin UI/API callers and attaches Principal to context.
func AdminAuth(cfg config.Config, oidc *OIDCVerifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfg.AuthDisabled {
				next.ServeHTTP(w, r.WithContext(WithPrincipal(r.Context(), localDevPrincipal)))
				return
			}

			principal, ok, err := authenticateAdmin(r.Context(), cfg, oidc, r)
			if err != nil {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid token"})
				return
			}
			if !ok {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "admin authentication required"})
				return
			}
			next.ServeHTTP(w, r.WithContext(WithPrincipal(r.Context(), principal)))
		})
	}
}

func authenticateAdmin(ctx context.Context, cfg config.Config, oidc *OIDCVerifier, r *http.Request) (Principal, bool, error) {
	bearer := bearerToken(r)
	if bearer != "" && cfg.OIDCEnabled && oidc != nil && looksLikeJWT(bearer) {
		p, err := oidc.Verify(ctx, bearer)
		if err == nil {
			return p, true, nil
		}
		return Principal{}, false, err
	}

	key := strings.TrimSpace(r.Header.Get(adminKeyHeader))
	if key == "" {
		key = bearer
	}
	if key == "" {
		return Principal{}, false, nil
	}

	if roles, ok := apiKeyRoles(cfg, key); ok {
		return Principal{
			Subject:    "api-key",
			Roles:      roles,
			AuthMethod: "api_key",
		}, true, nil
	}

	return Principal{}, false, nil
}

func bearerToken(r *http.Request) string {
	header := r.Header.Get("Authorization")
	if !strings.HasPrefix(header, "Bearer ") {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
}

func apiKeyRoles(cfg config.Config, key string) ([]Role, bool) {
	type keyRole struct {
		value string
		role  Role
	}
	candidates := []keyRole{
		{cfg.AdminAPIKey, RoleAdmin},
		{cfg.PublisherAPIKey, RolePublisher},
		{cfg.ReaderAPIKey, RoleReader},
	}
	for _, c := range candidates {
		if c.value == "" {
			continue
		}
		if subtle.ConstantTimeCompare([]byte(key), []byte(c.value)) == 1 {
			return []Role{c.role}, true
		}
	}
	return nil, false
}

// AdminKey is kept for backward compatibility in tests.
func AdminKey(cfg config.Config) func(http.Handler) http.Handler {
	return AdminAuth(cfg, nil)
}
