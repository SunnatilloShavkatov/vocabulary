package authcontroller

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"vocabulary/backend/modules/auth/service"
)

type authLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func RegisterAuthRoutes(mux *http.ServeMux, svc *authservice.AuthService, protected func(http.HandlerFunc) http.HandlerFunc) {
	h := &AuthHandler{service: svc}
	mux.HandleFunc("POST /v1/auth/login", h.login)
	mux.HandleFunc("POST /v1/admins", protected(h.createAdmin))
}

type AuthHandler struct {
	service *authservice.AuthService
}

func (h *AuthHandler) login(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		writeAuthJSON(w, http.StatusInternalServerError, map[string]string{"error": "service not initialized"})
		return
	}

	var req authLoginRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeAuthJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		writeAuthJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" || req.Password == "" {
		writeAuthJSON(w, http.StatusBadRequest, map[string]string{"error": "email and password are required"})
		return
	}

	resp, err := h.service.Login(authservice.AuthLoginRequest{Email: req.Email, Password: req.Password})
	if err != nil {
		switch {
		case errors.Is(err, authservice.ErrBootstrapAdminNotConfigured):
			writeAuthJSON(w, http.StatusServiceUnavailable, map[string]string{"error": err.Error()})
		case errors.Is(err, authservice.ErrInvalidCredentials):
			writeAuthJSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
		default:
			writeAuthJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create access token"})
		}
		return
	}

	writeAuthJSON(w, http.StatusOK, resp)
}

func (h *AuthHandler) createAdmin(w http.ResponseWriter, _ *http.Request) {
	writeAuthJSON(w, http.StatusNotImplemented, map[string]string{"module": "auth", "error": "not implemented"})
}

func writeAuthJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

