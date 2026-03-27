package auth

import (
	"fmt"
	"net/http"
	"strings"
)

// InternalAuth returns middleware that validates a bearer token for internal endpoints.
func InternalAuth(token string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if token == "" {
				// No token configured — reject all internal requests
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				fmt.Fprint(w, `{"error":{"code":"UNAUTHORIZED","message":"internal auth not configured"}}`)
				return
			}

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				fmt.Fprint(w, `{"error":{"code":"UNAUTHORIZED","message":"missing or invalid authorization header"}}`)
				return
			}

			provided := strings.TrimPrefix(authHeader, "Bearer ")
			if provided != token {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				fmt.Fprint(w, `{"error":{"code":"UNAUTHORIZED","message":"invalid token"}}`)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
