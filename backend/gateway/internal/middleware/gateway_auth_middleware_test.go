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
	handler := RequireGatewayAdmin(testSecret, nil)(func(w http.ResponseWriter, _ *http.Request) {
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
	handler := RequireGatewayAdmin(testSecret, nil)(func(w http.ResponseWriter, _ *http.Request) {
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
	handler := RequireGatewayAdmin(testSecret, nil)(func(w http.ResponseWriter, _ *http.Request) {
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
	handler := RequireGatewayAdmin(testSecret, nil)(func(w http.ResponseWriter, _ *http.Request) {
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
	handler := RequireGatewayAdmin(testSecret, nil)(func(w http.ResponseWriter, r *http.Request) {
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

func TestRequireGatewayAuth_CachedToken(t *testing.T) {
	cache := NewTokenCache(100)
	calledCount := 0

	handler := RequireGatewayAuth(testSecret, cache)(func(w http.ResponseWriter, r *http.Request) {
		calledCount++
		claims, ok := r.Context().Value(GatewayClaimsKey).(jwt.MapClaims)
		if !ok || claims["sub"] != "test" {
			t.Error("expected claims in context")
		}
		w.WriteHeader(http.StatusOK)
	})

	token := makeToken(t, "user", false)

	// 1st request (Cache Miss, parses JWT)
	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	req1.Header.Set("Authorization", "Bearer "+token)
	res1 := httptest.NewRecorder()
	handler(res1, req1)
	if res1.Code != http.StatusOK {
		t.Fatalf("first request failed: status %d", res1.Code)
	}

	// Verify it got cached
	if _, found := cache.Get(token); !found {
		t.Fatal("expected token to be cached")
	}

	// 2nd request (Cache Hit, should bypass parsing JWT and be very fast)
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.Header.Set("Authorization", "Bearer "+token)
	res2 := httptest.NewRecorder()
	handler(res2, req2)
	if res2.Code != http.StatusOK {
		t.Fatalf("second request failed: status %d", res2.Code)
	}

	if calledCount != 2 {
		t.Fatalf("expected handler to be called 2 times, got %d", calledCount)
	}
}

