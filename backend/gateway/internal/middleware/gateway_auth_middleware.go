package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	jwt "github.com/golang-jwt/jwt/v5"
)

type contextKey string

// GatewayClaimsKey is the context key under which validated JWT claims are stored.
const GatewayClaimsKey contextKey = "jwt_claims"

// RequireGatewayAdmin returns a middleware that validates a Bearer JWT and enforces role=admin.
func RequireGatewayAdmin(secret string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				writeGatewayJSON(w, http.StatusUnauthorized, map[string]string{"error": "authorization header required"})
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
				}
				return []byte(secret), nil
			})
			if err != nil || !token.Valid {
				writeGatewayJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid or expired token"})
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				writeGatewayJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid token claims"})
				return
			}

			if claims["role"] != "admin" {
				writeGatewayJSON(w, http.StatusForbidden, map[string]string{"error": "admin role required"})
				return
			}

			ctx := context.WithValue(r.Context(), GatewayClaimsKey, claims)
			next(w, r.WithContext(ctx))
		}
	}
}

func writeGatewayJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

