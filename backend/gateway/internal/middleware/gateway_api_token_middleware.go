package gatewaymiddleware

import (
	"encoding/json"
	"net/http"
	"strings"
)

// WithGatewayAPIToken enforces a static API token check on all requests except /healthz and /swagger/.
func WithGatewayAPIToken(expectedToken string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Bypass health check and swagger documentation endpoints
		if r.URL.Path == "/healthz" || strings.HasPrefix(r.URL.Path, "/swagger") {
			next.ServeHTTP(w, r)
			return
		}

		// Enforce API token check if configured
		if expectedToken != "" {
			token := r.Header.Get("X-API-Token")
			if token == "" || token != expectedToken {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid or missing API token"})
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
