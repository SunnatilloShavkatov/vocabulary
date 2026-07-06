package gatewayhttp

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"vocabulary/backend/gateway/internal/grpcclient"
	"vocabulary/backend/gateway/internal/middleware"
	"vocabulary/backend/modules/notification/controller"
	"vocabulary/backend/modules/notification/service"
	"vocabulary/backend/modules/users/controller"
	"vocabulary/backend/modules/users/service"
	"vocabulary/backend/modules/vocabulary/controller"
	"vocabulary/backend/modules/vocabulary/service"
)

type GatewayRouter struct {
	mux                *http.ServeMux
	corsAllowedOrigins string
	authGRPCClient     grpcclient.AuthServiceClient
}

func NewGatewayRouter(
	jwtSecret string,
	corsAllowedOrigins string,
	vocabularySvc *vocabularyservice.VocabularyService,
	usersSvc *usersservice.UsersService,
	notificationSvc *notificationservice.NotificationService,
	authGRPCClient grpcclient.AuthServiceClient,
) *GatewayRouter {
	mux := http.NewServeMux()
	r := &GatewayRouter{mux: mux, corsAllowedOrigins: corsAllowedOrigins, authGRPCClient: authGRPCClient}
	r.registerGatewayRoutes(jwtSecret, vocabularySvc, usersSvc, notificationSvc)
	return r
}

func (r *GatewayRouter) Handler() http.Handler {
	return gatewaymiddleware.WithGatewayCORS(r.corsAllowedOrigins, withGatewayRequestLogging(r.mux))
}

func (r *GatewayRouter) registerGatewayRoutes(
	jwtSecret string,
	vocabularySvc *vocabularyservice.VocabularyService,
	usersSvc *usersservice.UsersService,
	notificationSvc *notificationservice.NotificationService,
) {
	protectedAuth := gatewaymiddleware.RequireGatewayAuth(jwtSecret)
	protectedAdmin := gatewaymiddleware.RequireGatewayAdmin(jwtSecret)
	authProxy := newGatewayAuthProxyHandler(r.authGRPCClient)

	r.mux.HandleFunc("GET /healthz", GatewayHealthHandler)
	r.mux.HandleFunc("GET /metrics", GatewayMetricsHandler)
	r.mux.HandleFunc("GET /internal/grpc/auth/health", r.gatewayAuthGRPCHealthHandler)
	r.mux.HandleFunc("GET /v1", gatewayVersionHandler)

	r.mux.HandleFunc("POST /v1/auth/login", authProxy.login)
	r.mux.HandleFunc("POST /v1/admins", protectedAdmin(authProxy.createAdmin))
	vocabularycontroller.RegisterVocabularyRoutes(r.mux, vocabularySvc, protectedAuth, protectedAdmin, gatewaymiddleware.GetGatewayAdminSubject)
	userscontroller.RegisterUsersRoutes(r.mux, usersSvc, protectedAuth, protectedAdmin)
	notificationcontroller.RegisterNotificationRoutes(r.mux, notificationSvc, protectedAuth)
}

func (r *GatewayRouter) gatewayAuthGRPCHealthHandler(w http.ResponseWriter, req *http.Request) {
	if r.authGRPCClient == nil {
		writeGatewayJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "not_ready", "reason": "auth_grpc_client_not_configured"})
		return
	}

	if err := r.authGRPCClient.CheckConnection(context.Background()); err != nil {
		writeGatewayJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "not_ready", "reason": "auth_grpc_unreachable", "target": r.authGRPCClient.Target()})
		return
	}

	writeGatewayJSON(w, http.StatusOK, map[string]string{"status": "ready", "target": r.authGRPCClient.Target()})
}

func GatewayHealthHandler(w http.ResponseWriter, _ *http.Request) {
	writeGatewayJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func gatewayVersionHandler(w http.ResponseWriter, _ *http.Request) {
	writeGatewayJSON(w, http.StatusOK, map[string]string{"service": "gateway", "version": "0.3.0"})
}

func writeGatewayJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

type gatewayStatusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *gatewayStatusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func withGatewayRequestLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &gatewayStatusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		recordGatewayMetrics(rec.status)

		elapsed := time.Since(start)
		if rec.status >= http.StatusInternalServerError {
			log.Printf("gateway request error method=%s path=%s status=%d duration=%s", r.Method, r.URL.Path, rec.status, elapsed)
			return
		}
		log.Printf("gateway request method=%s path=%s status=%d duration=%s", r.Method, r.URL.Path, rec.status, elapsed)
	})
}

