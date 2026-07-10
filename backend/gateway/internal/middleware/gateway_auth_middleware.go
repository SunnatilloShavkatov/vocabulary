package gatewaymiddleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

type contextKey string

// GatewayClaimsKey is the context key under which validated JWT claims are stored.
const GatewayClaimsKey contextKey = "jwt_claims"

const (
	GatewayUserIDHeader   = "X-User-ID"
	GatewayUserRoleHeader = "X-User-Role"
)

// RequireGatewayAuth returns a middleware that validates a Bearer JWT.
func RequireGatewayAuth(secret string, cache *TokenCache) func(http.HandlerFunc) http.HandlerFunc {
	return requireGatewayRole(secret, "", cache)
}

// RequireGatewayAdmin returns a middleware that validates a Bearer JWT and enforces role=admin.
func RequireGatewayAdmin(secret string, cache *TokenCache) func(http.HandlerFunc) http.HandlerFunc {
	return requireGatewayRole(secret, "admin", cache)
}

func requireGatewayRole(secret, requiredRole string, cache *TokenCache) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				writeGatewayJSON(w, http.StatusUnauthorized, map[string]string{"error": "authorization header required"})
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

			var claims jwt.MapClaims
			var cacheHit bool

			if cache != nil {
				if cachedClaims, found := cache.Get(tokenStr); found {
					claims = cachedClaims
					cacheHit = true
				}
			}

			if !cacheHit {
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

				var ok bool
				claims, ok = token.Claims.(jwt.MapClaims)
				if !ok {
					writeGatewayJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid token claims"})
					return
				}

				if cache != nil {
					// Default TTL is 5 minutes
					ttl := 5 * time.Minute
					if expVal, exists := claims["exp"]; exists {
						if expFloat, ok := expVal.(float64); ok {
							expTime := time.Unix(int64(expFloat), 0)
							remaining := time.Until(expTime)
							if remaining > 0 {
								if remaining < ttl {
									ttl = remaining
								}
							} else {
								writeGatewayJSON(w, http.StatusUnauthorized, map[string]string{"error": "token is expired"})
								return
							}
						}
					}
					cache.Set(tokenStr, claims, ttl)
				}
			}

			sub, subOK := claims["sub"].(string)
			role, roleOK := claims["role"].(string)
			if strings.TrimSpace(sub) == "" || strings.TrimSpace(role) == "" || !subOK || !roleOK {
				writeGatewayJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid token claims"})
				return
			}

			if requiredRole != "" && role != requiredRole {
				writeGatewayJSON(w, http.StatusForbidden, map[string]string{"error": "admin role required"})
				return
			}

			ctx := context.WithValue(r.Context(), GatewayClaimsKey, claims)
			r = r.WithContext(ctx)
			r.Header.Set(GatewayUserIDHeader, strings.TrimSpace(sub))
			r.Header.Set(GatewayUserRoleHeader, strings.TrimSpace(role))
			next(w, r)
		}
	}
}

func GetGatewayAdminSubject(ctx context.Context) (string, bool) {
	claims, ok := ctx.Value(GatewayClaimsKey).(jwt.MapClaims)
	if !ok {
		return "", false
	}
	sub, ok := claims["sub"].(string)
	if !ok || strings.TrimSpace(sub) == "" {
		return "", false
	}
	return strings.TrimSpace(sub), true
}

func GetGatewayIdentityFromHeaders(r *http.Request) (string, string, bool) {
	if r == nil {
		return "", "", false
	}
	userID := strings.TrimSpace(r.Header.Get(GatewayUserIDHeader))
	role := strings.TrimSpace(r.Header.Get(GatewayUserRoleHeader))
	if userID == "" || role == "" {
		return "", "", false
	}
	return userID, role, true
}

func writeGatewayJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

