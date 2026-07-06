package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"vocabulary/backend/libs/shared/config"
	"vocabulary/backend/libs/shared/db"
	authcontroller "vocabulary/backend/modules/auth/controller"
	authrepository "vocabulary/backend/modules/auth/repository"
	authservice "vocabulary/backend/modules/auth/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	pool, err := db.NewPool(context.Background(), cfg.Database)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	repo := authrepository.NewAuthPgxRepository(pool)
	svc := authservice.NewAuthService(cfg, repo)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", authHealthHandler)
	mux.HandleFunc("GET /readyz", authReadyHandler(pool))
	mux.HandleFunc("GET /metrics", authMetricsHandler)
	authcontroller.RegisterAuthRoutes(mux, svc, requireAuthAdmin(cfg.JWT.Secret))
	handler := withAuthMetrics(mux)

	port := authServerPort(cfg.Server.Port)
	addr := fmt.Sprintf(":%d", port)
	log.Printf("auth-service listening on %s", addr)

	errCh := make(chan error, 2)
	go func() {
		errCh <- http.ListenAndServe(addr, handler)
	}()

	grpcPort := authGRPCPort(9091)
	log.Printf("auth-service grpc listening on :%d", grpcPort)
	go func() {
		errCh <- startAuthGRPCServer(grpcPort, svc)
	}()

	err = <-errCh
	log.Fatalf("auth-service stopped: %v", err)
}

func authServerPort(defaultPort int) int {
	if defaultPort <= 0 {
		defaultPort = 8081
	}
	v := strings.TrimSpace(os.Getenv("AUTH_SERVICE_PORT"))
	if v == "" {
		return defaultPort
	}
	p, err := strconv.Atoi(v)
	if err != nil || p <= 0 {
		return defaultPort
	}
	return p
}

func authGRPCPort(defaultPort int) int {
	v := strings.TrimSpace(os.Getenv("AUTH_GRPC_PORT"))
	if v == "" {
		return defaultPort
	}
	p, err := strconv.Atoi(v)
	if err != nil || p <= 0 {
		return defaultPort
	}
	return p
}

func authHealthHandler(w http.ResponseWriter, _ *http.Request) {
	writeAuthServiceJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "auth-service"})
}

func authReadyHandler(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			writeAuthServiceJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "not_ready", "reason": "database_unreachable"})
			return
		}

		writeAuthServiceJSON(w, http.StatusOK, map[string]string{"status": "ready", "service": "auth-service"})
	}
}

func requireAuthAdmin(secret string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				writeAuthServiceJSON(w, http.StatusUnauthorized, map[string]string{"error": "authorization header required"})
				return
			}

			tokenStr := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
			token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
				return []byte(secret), nil
			})
			if err != nil || !token.Valid {
				writeAuthServiceJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid or expired token"})
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				writeAuthServiceJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid token claims"})
				return
			}
			if strings.TrimSpace(fmt.Sprint(claims["role"])) != "admin" {
				writeAuthServiceJSON(w, http.StatusForbidden, map[string]string{"error": "admin role required"})
				return
			}

			next(w, r)
		}
	}
}

func writeAuthServiceJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
