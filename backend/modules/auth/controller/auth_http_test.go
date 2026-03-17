package authcontroller

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"vocabulary/backend/libs/shared/config"
	"vocabulary/backend/modules/auth/service"
)

func noopProtected(next http.HandlerFunc) http.HandlerFunc { return next }

type mockAuthRepository struct {
	admins map[string]authservice.AuthAdmin
}

func (m *mockAuthRepository) CreateAdmin(_ context.Context, email, _ string, role string) (*authservice.AuthAdmin, error) {
	if m.admins == nil {
		m.admins = map[string]authservice.AuthAdmin{}
	}
	if _, ok := m.admins[email]; ok {
		return nil, authservice.ErrAuthAdminAlreadyExists
	}
	admin := authservice.AuthAdmin{
		ID:        "admin-id-1",
		Email:     email,
		Role:      role,
		CreatedAt: time.Now().UTC(),
	}
	m.admins[email] = admin
	return &admin, nil
}

func TestLoginSuccess(t *testing.T) {
	mux := http.NewServeMux()
	svc := authservice.NewAuthService(config.Config{
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

	var response authservice.AuthLoginResponse
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
	RegisterAuthRoutes(mux, authservice.NewAuthService(config.Config{}), noopProtected)

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

func TestCreateAdminSuccess(t *testing.T) {
	mux := http.NewServeMux()
	repo := &mockAuthRepository{}
	RegisterAuthRoutes(mux, authservice.NewAuthService(config.Config{}, repo), noopProtected)
	req := httptest.NewRequest(http.MethodPost, "/v1/admins", bytes.NewBufferString(`{"email":"new-admin@example.com","password":"password123"}`))
	res := httptest.NewRecorder()
	mux.ServeHTTP(res, req)
	if res.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, res.Code, res.Body.String())
	}
}

func TestCreateAdminValidationError(t *testing.T) {
	mux := http.NewServeMux()
	RegisterAuthRoutes(mux, authservice.NewAuthService(config.Config{}, &mockAuthRepository{}), noopProtected)
	req := httptest.NewRequest(http.MethodPost, "/v1/admins", bytes.NewBufferString(`{"email":"new-admin@example.com","password":"short"}`))
	res := httptest.NewRecorder()
	mux.ServeHTTP(res, req)
	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, res.Code)
	}
}

func TestCreateAdminDuplicate(t *testing.T) {
	mux := http.NewServeMux()
	repo := &mockAuthRepository{}
	RegisterAuthRoutes(mux, authservice.NewAuthService(config.Config{}, repo), noopProtected)
	body := `{"email":"dup-admin@example.com","password":"password123"}`

	firstReq := httptest.NewRequest(http.MethodPost, "/v1/admins", bytes.NewBufferString(body))
	firstRes := httptest.NewRecorder()
	mux.ServeHTTP(firstRes, firstReq)
	if firstRes.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, firstRes.Code)
	}

	secondReq := httptest.NewRequest(http.MethodPost, "/v1/admins", bytes.NewBufferString(body))
	secondRes := httptest.NewRecorder()
	mux.ServeHTTP(secondRes, secondReq)
	if secondRes.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, secondRes.Code)
	}
}

