package vocabularyservice

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	"vocabulary/backend/libs/shared/config"
)

type VocabularyItem struct {
	ID          string    `json:"id"`
	Word        string    `json:"word"`
	Translation string    `json:"translation"`
	Example     string    `json:"example,omitempty"`
	Category    string    `json:"category,omitempty"`
	Status      string    `json:"status,omitempty"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type AdminVocabularySearch struct {
	Word        string
	Translation string
	Category    string
	Status      string
	Page        int
	Limit       int
}

type AdminVocabularyUpsertInput struct {
	Word        string
	Translation string
	Example     string
	Category    string
	Status      string
	CreatedBy   string
}

type VocabularyRepository interface {
	Create(ctx context.Context, word, translation, example string, createdBy *string) (*VocabularyItem, error)
	List(ctx context.Context, search string, page, limit int) ([]VocabularyItem, int, error)
	AdminList(ctx context.Context, filter AdminVocabularySearch) ([]VocabularyItem, int, error)
	AdminCreate(ctx context.Context, input AdminVocabularyUpsertInput) (*VocabularyItem, error)
	AdminGet(ctx context.Context, id string) (*VocabularyItem, error)
	AdminUpdate(ctx context.Context, id string, input AdminVocabularyUpsertInput) (*VocabularyItem, error)
	AdminDelete(ctx context.Context, id string) error
	AdminSetStatus(ctx context.Context, id string, status string) (*VocabularyItem, error)
}

type VocabularyService struct {
	repo VocabularyRepository
}

var vocabularyUUIDLikeRegexp = regexp.MustCompile(`^[0-9a-fA-F-]{36}$`)

func NewVocabularyService(_ config.Config, repo VocabularyRepository) *VocabularyService {
	return &VocabularyService{repo: repo}
}

func (s *VocabularyService) Create(ctx context.Context, word, translation, example, createdBy string) (*VocabularyItem, error) {
	normalizedCreatedBy := normalizeVocabularyCreatedBy(createdBy)
	return s.repo.Create(ctx, strings.TrimSpace(word), strings.TrimSpace(translation), strings.TrimSpace(example), normalizedCreatedBy)
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

func (s *VocabularyService) AdminList(ctx context.Context, filter AdminVocabularySearch) ([]VocabularyItem, int, error) {
	filter.Word = strings.TrimSpace(filter.Word)
	filter.Translation = strings.TrimSpace(filter.Translation)
	filter.Category = strings.TrimSpace(filter.Category)
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 || filter.Limit > 100 {
		filter.Limit = 20
	}
	if filter.Status != "" {
		status, err := normalizeVocabularyStatus(filter.Status)
		if err != nil {
			return nil, 0, err
		}
		filter.Status = status
	}
	return s.repo.AdminList(ctx, filter)
}

func (s *VocabularyService) AdminCreate(ctx context.Context, input AdminVocabularyUpsertInput) (*VocabularyItem, error) {
	normalized, err := normalizeAdminVocabularyInput(input, true)
	if err != nil {
		return nil, err
	}
	return s.repo.AdminCreate(ctx, normalized)
}

func (s *VocabularyService) AdminGet(ctx context.Context, id string) (*VocabularyItem, error) {
	if strings.TrimSpace(id) == "" {
		return nil, ErrVocabularyNotFound
	}
	return s.repo.AdminGet(ctx, strings.TrimSpace(id))
}

func (s *VocabularyService) AdminUpdate(ctx context.Context, id string, input AdminVocabularyUpsertInput) (*VocabularyItem, error) {
	if strings.TrimSpace(id) == "" {
		return nil, ErrVocabularyNotFound
	}
	normalized, err := normalizeAdminVocabularyInput(input, false)
	if err != nil {
		return nil, err
	}
	return s.repo.AdminUpdate(ctx, strings.TrimSpace(id), normalized)
}

func (s *VocabularyService) AdminDelete(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return ErrVocabularyNotFound
	}
	return s.repo.AdminDelete(ctx, strings.TrimSpace(id))
}

func (s *VocabularyService) AdminApprove(ctx context.Context, id string) (*VocabularyItem, error) {
	if strings.TrimSpace(id) == "" {
		return nil, ErrVocabularyNotFound
	}
	return s.repo.AdminSetStatus(ctx, strings.TrimSpace(id), "approved")
}

func (s *VocabularyService) AdminReject(ctx context.Context, id string) (*VocabularyItem, error) {
	if strings.TrimSpace(id) == "" {
		return nil, ErrVocabularyNotFound
	}
	return s.repo.AdminSetStatus(ctx, strings.TrimSpace(id), "rejected")
}

func normalizeVocabularyCreatedBy(createdBy string) *string {
	v := strings.TrimSpace(createdBy)
	if v == "" {
		return nil
	}
	if !vocabularyUUIDLikeRegexp.MatchString(v) {
		return nil
	}
	return &v
}

var (
	ErrInvalidVocabularyInput = errors.New("invalid vocabulary input")
	ErrVocabularyNotFound     = errors.New("vocabulary not found")
	ErrInvalidVocabularyStatus = errors.New("invalid vocabulary status")
)

func normalizeAdminVocabularyInput(input AdminVocabularyUpsertInput, defaultApproved bool) (AdminVocabularyUpsertInput, error) {
	input.Word = strings.TrimSpace(input.Word)
	input.Translation = strings.TrimSpace(input.Translation)
	input.Example = strings.TrimSpace(input.Example)
	input.Category = strings.TrimSpace(input.Category)
	input.CreatedBy = strings.TrimSpace(input.CreatedBy)

	if input.Word == "" || input.Translation == "" {
		return AdminVocabularyUpsertInput{}, ErrInvalidVocabularyInput
	}

	if strings.TrimSpace(input.Status) == "" && defaultApproved {
		input.Status = "approved"
	}
	if strings.TrimSpace(input.Status) == "" && !defaultApproved {
		input.Status = ""
		return input, nil
	}

	status, err := normalizeVocabularyStatus(input.Status)
	if err != nil {
		return AdminVocabularyUpsertInput{}, err
	}
	input.Status = status
	return input, nil
}

func normalizeVocabularyStatus(status string) (string, error) {
	v := strings.ToLower(strings.TrimSpace(status))
	switch v {
	case "", "pending", "approved", "rejected", "blocked":
		if v == "" {
			return "pending", nil
		}
		return v, nil
	default:
		return "", ErrInvalidVocabularyStatus
	}
}

