package gatewaymiddleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWithGatewayAPIToken(t *testing.T) {
	const expectedToken = "secret-token-123"

	handler := WithGatewayAPIToken(expectedToken, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))

	t.Run("Valid token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/vocabulary", nil)
		req.Header.Set("X-API-Token", expectedToken)
		res := httptest.NewRecorder()
		handler.ServeHTTP(res, req)
		if res.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", res.Code)
		}
	})

	t.Run("Missing token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/vocabulary", nil)
		res := httptest.NewRecorder()
		handler.ServeHTTP(res, req)
		if res.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", res.Code)
		}
		var errResp map[string]string
		_ = json.Unmarshal(res.Body.Bytes(), &errResp)
		if errResp["error"] != "invalid or missing API token" {
			t.Errorf("expected 'invalid or missing API token', got '%s'", errResp["error"])
		}
	})

	t.Run("Invalid token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/vocabulary", nil)
		req.Header.Set("X-API-Token", "wrong-token")
		res := httptest.NewRecorder()
		handler.ServeHTTP(res, req)
		if res.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", res.Code)
		}
	})

	t.Run("Bypass health check", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		res := httptest.NewRecorder()
		handler.ServeHTTP(res, req)
		if res.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", res.Code)
		}
	})

	t.Run("Bypass swagger UI", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/swagger/", nil)
		res := httptest.NewRecorder()
		handler.ServeHTTP(res, req)
		if res.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", res.Code)
		}
	})
}
