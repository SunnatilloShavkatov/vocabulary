package gatewaymiddleware

import "net/http"

// WithGatewayCORS wraps a handler with permissive CORS headers for local development.
// In production, restrict AllowedOrigins to the actual admin web origin.
func WithGatewayCORS(allowedOrigins string, next http.Handler) http.Handler {
	if allowedOrigins == "" {
		allowedOrigins = "*"
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", allowedOrigins)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "86400")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

