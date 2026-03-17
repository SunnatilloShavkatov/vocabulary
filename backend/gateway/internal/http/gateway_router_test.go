package gatewayhttp

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"vocabulary/backend/libs/shared/config"
	"vocabulary/backend/modules/auth/service"
	"vocabulary/backend/modules/vocabulary/service"
)

// mockVocabularyRepo satisfies vocabulary.VocabularyRepository without a real DB.
type mockVocabularyRepo struct{}

func (m *mockVocabularyRepo) Create(_ context.Context, word, translation, _ string) (*vocabularyservice.VocabularyItem, error) {
	return &vocabularyservice.VocabularyItem{ID: "test-id", Word: word, Translation: translation}, nil
}

func (m *mockVocabularyRepo) List(_ context.Context, _ string, _, _ int) ([]vocabularyservice.VocabularyItem, int, error) {
	return []vocabularyservice.VocabularyItem{}, 0, nil
}

func TestHealthHandler(t *testing.T) {
	r := newTestRouter()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	res := httptest.NewRecorder()
	r.Handler().ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, res.Code)
	}
}

func TestVersionRoute(t *testing.T) {
	r := newTestRouter()
	req := httptest.NewRequest(http.MethodGet, "/v1", nil)
	res := httptest.NewRecorder()
	r.Handler().ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, res.Code)
	}
}

func TestModuleRoutes(t *testing.T) {
	r := newTestRouter()

	tests := []struct {
		name   string
		method string
		path   string
		body   string
		status int
	}{
		{
			name:   "auth login success",
			method: http.MethodPost,
			path:   "/v1/auth/login",
			body:   `{"email":"admin@example.com","password":"password123"}`,
			status: http.StatusOK,
		},
		{
			name:   "create admin - no auth header",
			method: http.MethodPost,
			path:   "/v1/admins",
			status: http.StatusUnauthorized,
		},
		{
			name:   "create vocabulary - no auth header",
			method: http.MethodPost,
			path:   "/v1/vocabulary",
			status: http.StatusUnauthorized,
		},
		{
			name:   "list vocabulary - public",
			method: http.MethodGet,
			path:   "/v1/vocabulary",
			status: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, bytes.NewBufferString(tt.body))
			res := httptest.NewRecorder()
			r.Handler().ServeHTTP(res, req)
			if res.Code != tt.status {
				t.Fatalf("expected status %d, got %d: %s", tt.status, res.Code, res.Body.String())
			}
		})
	}
}

func newTestRouter() *GatewayRouter {
	cfg := config.Config{
		JWT: config.JWTConfig{
			Secret:           "test-secret",
			AccessTTLMinutes: 15,
		},
		BootstrapAdmin: config.BootstrapAdminConfig{
			Email:    "admin@example.com",
			Password: "password123",
		},
	}
	authSvc := authservice.NewAuthService(cfg)
	vocabularySvc := vocabularyservice.NewVocabularyService(cfg, &mockVocabularyRepo{})
	return NewGatewayRouter("test-secret", authSvc, vocabularySvc)
}

