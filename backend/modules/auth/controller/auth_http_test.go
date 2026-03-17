package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	jwt "github.com/golang-jwt/jwt/v5"
	"vocabulary/backend/libs/shared/config"
	"vocabulary/backend/modules/auth/service"
)

func noopProtected(next http.HandlerFunc) http.HandlerFunc { return next }

func TestLoginSuccess(t *testing.T) {
	mux := http.NewServeMux()
	svc := service.NewAuthService(config.Config{
		JWT: config.JWTConfig{Secret: "test-secret", AccessTTLMinutes: 5},
		BootstrapAdmin: config.BootstrapAdminConfig{Email: "admin@example.com", Password: "password123"},
	})
	RegisterAuthRoutes(mux, svc, noopProtected)

	req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewBufferString(`{"email":"admin@example.com","password":"password123"}`))
	res := httptest.NewRecorder()
	mux.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, res.Code)
	}

	var response service.AuthLoginResponse
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		t.Fatalf("expected valid json response, got %v", err)
	}

	parsedToken, err := jwt.Parse(response.AccessToken, func(token *jwt.Token) (any, error) {
		return []byte("test-secret"), nil
	})
	if err != nil || !parsedToken.Valid {
		t.Fatalf("expected valid jwt token, got err=%v", err)
	}
}

func TestLoginValidationErrors(t *testing.T) {
	mux := http.NewServeMux()
	RegisterAuthRoutes(mux, service.NewAuthService(config.Config{}), noopProtected)

	tests := []string{`{"email":`, `{}`, `{"email":"a","password":"b"} {}`}
	for _, body := range tests {
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewBufferString(body))
		res := httptest.NewRecorder()
		mux.ServeHTTP(res, req)
		if res.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, res.Code)
		}
	}
}

func TestCreateAdminNotImplemented(t *testing.T) {
	mux := http.NewServeMux()
	RegisterAuthRoutes(mux, service.NewAuthService(config.Config{}), noopProtected)
	req := httptest.NewRequest(http.MethodPost, "/v1/admins", nil)
	res := httptest.NewRecorder()
	mux.ServeHTTP(res, req)
	if res.Code != http.StatusNotImplemented {
		t.Fatalf("expected status %d, got %d", http.StatusNotImplemented, res.Code)
	}
}

