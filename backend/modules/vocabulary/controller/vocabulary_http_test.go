package vocabularycontroller

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"vocabulary/backend/libs/shared/config"
	"vocabulary/backend/modules/vocabulary/service"
)

func noopProtected(next http.HandlerFunc) http.HandlerFunc { return next }

type mockRepository struct {
	items []vocabularyservice.VocabularyItem
}

func (m *mockRepository) Create(_ context.Context, word, translation, example string, _ *string) (*vocabularyservice.VocabularyItem, error) {
	item := vocabularyservice.VocabularyItem{ID: "test-id", Word: word, Translation: translation, Example: example, CreatedAt: time.Now()}
	m.items = append(m.items, item)
	return &item, nil
}

func (m *mockRepository) List(_ context.Context, _ string, _, _ int) ([]vocabularyservice.VocabularyItem, int, error) {
	return m.items, len(m.items), nil
}

func TestCreateVocabularySuccess(t *testing.T) {
	mux := http.NewServeMux()
	RegisterVocabularyRoutes(mux, vocabularyservice.NewVocabularyService(config.Config{}, &mockRepository{}), noopProtected, nil)
	req := httptest.NewRequest(http.MethodPost, "/v1/vocabulary", bytes.NewBufferString(`{"word":"apple","translation":"olma"}`))
	res := httptest.NewRecorder()
	mux.ServeHTTP(res, req)
	if res.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, res.Code)
	}
}

func TestCreateVocabularyValidation(t *testing.T) {
	mux := http.NewServeMux()
	RegisterVocabularyRoutes(mux, vocabularyservice.NewVocabularyService(config.Config{}, &mockRepository{}), noopProtected, nil)
	req := httptest.NewRequest(http.MethodPost, "/v1/vocabulary", bytes.NewBufferString(`{"word":""}`))
	res := httptest.NewRecorder()
	mux.ServeHTTP(res, req)
	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, res.Code)
	}
}

func TestListVocabulary(t *testing.T) {
	repo := &mockRepository{items: []vocabularyservice.VocabularyItem{{ID: "1", Word: "apple", Translation: "olma"}}}
	mux := http.NewServeMux()
	RegisterVocabularyRoutes(mux, vocabularyservice.NewVocabularyService(config.Config{}, repo), noopProtected, nil)
	req := httptest.NewRequest(http.MethodGet, "/v1/vocabulary?page=2&limit=10", nil)
	res := httptest.NewRecorder()
	mux.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, res.Code)
	}
	var resp vocabularyListResponse
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Meta.Page != 2 || resp.Meta.Limit != 10 || resp.Meta.Total != 1 {
		t.Fatalf("unexpected meta: %+v", resp.Meta)
	}
}

