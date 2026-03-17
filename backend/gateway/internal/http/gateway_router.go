package gatewayhttp

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"vocabulary/backend/gateway/internal/middleware"
	"vocabulary/backend/modules/auth/controller"
	"vocabulary/backend/modules/auth/service"
	"vocabulary/backend/modules/vocabulary/controller"
	"vocabulary/backend/modules/vocabulary/service"
)

type GatewayRouter struct {
	mux                *http.ServeMux
	corsAllowedOrigins string
}

func NewGatewayRouter(jwtSecret string, corsAllowedOrigins string, authSvc *authservice.AuthService, vocabularySvc *vocabularyservice.VocabularyService) *GatewayRouter {
	mux := http.NewServeMux()
	r := &GatewayRouter{mux: mux, corsAllowedOrigins: corsAllowedOrigins}
	r.registerGatewayRoutes(jwtSecret, authSvc, vocabularySvc)
	return r
}

func (r *GatewayRouter) Handler() http.Handler {
	return gatewaymiddleware.WithGatewayCORS(r.corsAllowedOrigins, withGatewayRequestLogging(r.mux))
}

func (r *GatewayRouter) registerGatewayRoutes(jwtSecret string, authSvc *authservice.AuthService, vocabularySvc *vocabularyservice.VocabularyService) {
	protected := gatewaymiddleware.RequireGatewayAdmin(jwtSecret)

	r.mux.HandleFunc("GET /healthz", GatewayHealthHandler)
	r.mux.HandleFunc("GET /v1", gatewayVersionHandler)

	authcontroller.RegisterAuthRoutes(r.mux, authSvc, protected)
	vocabularycontroller.RegisterVocabularyRoutes(r.mux, vocabularySvc, protected, gatewaymiddleware.GetGatewayAdminSubject)
}

func GatewayHealthHandler(w http.ResponseWriter, _ *http.Request) {
	writeGatewayJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func gatewayVersionHandler(w http.ResponseWriter, _ *http.Request) {
	writeGatewayJSON(w, http.StatusOK, map[string]string{"service": "gateway", "version": "0.1.0"})
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

		elapsed := time.Since(start)
		if rec.status >= http.StatusInternalServerError {
			log.Printf("gateway request error method=%s path=%s status=%d duration=%s", r.Method, r.URL.Path, rec.status, elapsed)
			return
		}
		log.Printf("gateway request method=%s path=%s status=%d duration=%s", r.Method, r.URL.Path, rec.status, elapsed)
	})
}

