package service

import (
	"context"
	"strings"
	"time"

	"vocabulary/backend/libs/shared/config"
)

type VocabularyItem struct {
	ID          string    `json:"id"`
	Word        string    `json:"word"`
	Translation string    `json:"translation"`
	Example     string    `json:"example,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type VocabularyRepository interface {
	Create(ctx context.Context, word, translation, example string) (*VocabularyItem, error)
	List(ctx context.Context, search string, page, limit int) ([]VocabularyItem, int, error)
}

type VocabularyService struct {
	repo VocabularyRepository
}

func NewVocabularyService(_ config.Config, repo VocabularyRepository) *VocabularyService {
	return &VocabularyService{repo: repo}
}

func (s *VocabularyService) Create(ctx context.Context, word, translation, example string) (*VocabularyItem, error) {
	return s.repo.Create(ctx, strings.TrimSpace(word), strings.TrimSpace(translation), strings.TrimSpace(example))
}

func (s *VocabularyService) List(ctx context.Context, search string, page, limit int) ([]VocabularyItem, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return s.repo.List(ctx, strings.TrimSpace(search), page, limit)
}

