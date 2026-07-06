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
func noopAdminProtected(next http.HandlerFunc) http.HandlerFunc { return next }

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

func (m *mockRepository) AdminList(_ context.Context, _ vocabularyservice.AdminVocabularySearch) ([]vocabularyservice.VocabularyItem, int, error) {
	return m.items, len(m.items), nil
}

func (m *mockRepository) AdminCreate(_ context.Context, input vocabularyservice.AdminVocabularyUpsertInput) (*vocabularyservice.VocabularyItem, error) {
	item := vocabularyservice.VocabularyItem{
		ID:          "admin-test-id",
		Word:        input.Word,
		Translation: input.Translation,
		Example:     input.Example,
		Category:    input.Category,
		Status:      input.Status,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	m.items = append(m.items, item)
	return &item, nil
}

func (m *mockRepository) AdminGet(_ context.Context, id string) (*vocabularyservice.VocabularyItem, error) {
	for _, item := range m.items {
		if item.ID == id {
			found := item
			return &found, nil
		}
	}
	return nil, vocabularyservice.ErrVocabularyNotFound
}

func (m *mockRepository) AdminUpdate(_ context.Context, id string, input vocabularyservice.AdminVocabularyUpsertInput) (*vocabularyservice.VocabularyItem, error) {
	for i := range m.items {
		if m.items[i].ID == id {
			m.items[i].Word = input.Word
			m.items[i].Translation = input.Translation
			m.items[i].Example = input.Example
			m.items[i].Category = input.Category
			m.items[i].Status = input.Status
			m.items[i].UpdatedAt = time.Now()
			updated := m.items[i]
			return &updated, nil
		}
	}
	return nil, vocabularyservice.ErrVocabularyNotFound
}

func (m *mockRepository) AdminDelete(_ context.Context, id string) error {
	for i := range m.items {
		if m.items[i].ID == id {
			m.items = append(m.items[:i], m.items[i+1:]...)
			return nil
		}
	}
	return vocabularyservice.ErrVocabularyNotFound
}

func (m *mockRepository) AdminSetStatus(_ context.Context, id string, status string) (*vocabularyservice.VocabularyItem, error) {
	for i := range m.items {
		if m.items[i].ID == id {
			m.items[i].Status = status
			m.items[i].UpdatedAt = time.Now()
			updated := m.items[i]
			return &updated, nil
		}
	}
	return nil, vocabularyservice.ErrVocabularyNotFound
}

func TestCreateVocabularySuccess(t *testing.T) {
	mux := http.NewServeMux()
	RegisterVocabularyRoutes(mux, vocabularyservice.NewVocabularyService(config.Config{}, &mockRepository{}), noopProtected, noopAdminProtected, nil)
	req := httptest.NewRequest(http.MethodPost, "/v1/vocabulary", bytes.NewBufferString(`{"word":"apple","translation":"olma"}`))
	res := httptest.NewRecorder()
	mux.ServeHTTP(res, req)
	if res.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, res.Code)
	}
}

func TestCreateVocabularyValidation(t *testing.T) {
	mux := http.NewServeMux()
	RegisterVocabularyRoutes(mux, vocabularyservice.NewVocabularyService(config.Config{}, &mockRepository{}), noopProtected, noopAdminProtected, nil)
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
	RegisterVocabularyRoutes(mux, vocabularyservice.NewVocabularyService(config.Config{}, repo), noopProtected, noopAdminProtected, nil)
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

