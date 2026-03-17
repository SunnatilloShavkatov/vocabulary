package gatewayhttp

import (
	"encoding/json"
	"net/http"

	"vocabulary/backend/gateway/internal/middleware"
	authcontroller "vocabulary/backend/modules/auth/controller"
	authservice "vocabulary/backend/modules/auth/service"
	vocabularycontroller "vocabulary/backend/modules/vocabulary/controller"
	vocabularyservice "vocabulary/backend/modules/vocabulary/service"
)

type GatewayRouter struct {
	mux *http.ServeMux
}

func NewGatewayRouter(jwtSecret string, authSvc *authservice.AuthService, vocabularySvc *vocabularyservice.VocabularyService) *GatewayRouter {
	mux := http.NewServeMux()
	r := &GatewayRouter{mux: mux}
	r.registerGatewayRoutes(jwtSecret, authSvc, vocabularySvc)
	return r
}

func (r *GatewayRouter) Handler() http.Handler {
	return r.mux
}

func (r *GatewayRouter) registerGatewayRoutes(jwtSecret string, authSvc *authservice.AuthService, vocabularySvc *vocabularyservice.VocabularyService) {
	protected := middleware.RequireGatewayAdmin(jwtSecret)

	r.mux.HandleFunc("GET /healthz", GatewayHealthHandler)
	r.mux.HandleFunc("GET /v1", gatewayVersionHandler)

	authcontroller.RegisterAuthRoutes(r.mux, authSvc, protected)
	vocabularycontroller.RegisterVocabularyRoutes(r.mux, vocabularySvc, protected)
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

