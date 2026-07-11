package auth

import "net/http"

func RequireRole(min Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			p, ok := PrincipalFromContext(r.Context())
			if !ok || !p.Has(min) {
				writeJSON(w, http.StatusForbidden, map[string]string{
					"error": "insufficient role",
				})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
