package gatewaymiddleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

const testSecret = "test-secret"

func makeToken(t *testing.T, role string, expired bool) string {
	t.Helper()
	now := time.Now()
	expiry := now.Add(15 * time.Minute)
	if expired {
		expiry = now.Add(-time.Minute)
	}
	claims := jwt.MapClaims{
		"sub":  "test",
		"role": role,
		"iat":  now.Unix(),
		"exp":  expiry.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(testSecret))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return signed
}

func TestRequireGatewayAdmin_NoHeader(t *testing.T) {
	handler := RequireGatewayAdmin(testSecret)(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	res := httptest.NewRecorder()
	handler(res, req)
	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected %d got %d", http.StatusUnauthorized, res.Code)
	}
}

func TestRequireGatewayAdmin_InvalidToken(t *testing.T) {
	handler := RequireGatewayAdmin(testSecret)(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Authorization", "Bearer bad-token")
	res := httptest.NewRecorder()
	handler(res, req)
	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected %d got %d", http.StatusUnauthorized, res.Code)
	}
}

func TestRequireGatewayAdmin_ExpiredToken(t *testing.T) {
	handler := RequireGatewayAdmin(testSecret)(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Authorization", "Bearer "+makeToken(t, "admin", true))
	res := httptest.NewRecorder()
	handler(res, req)
	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected %d got %d", http.StatusUnauthorized, res.Code)
	}
}

func TestRequireGatewayAdmin_NonAdminRole(t *testing.T) {
	handler := RequireGatewayAdmin(testSecret)(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Authorization", "Bearer "+makeToken(t, "user", false))
	res := httptest.NewRecorder()
	handler(res, req)
	if res.Code != http.StatusForbidden {
		t.Fatalf("expected %d got %d", http.StatusForbidden, res.Code)
	}
}

func TestRequireGatewayAdmin_ValidAdmin(t *testing.T) {
	called := false
	handler := RequireGatewayAdmin(testSecret)(func(w http.ResponseWriter, r *http.Request) {
		called = true
		claims, ok := r.Context().Value(GatewayClaimsKey).(jwt.MapClaims)
		if !ok || claims["role"] != "admin" {
			t.Error("expected admin claims in context")
		}
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Authorization", "Bearer "+makeToken(t, "admin", false))
	res := httptest.NewRecorder()
	handler(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("expected %d got %d", http.StatusOK, res.Code)
	}
	if !called {
		t.Fatal("expected next handler to be called")
	}
}

