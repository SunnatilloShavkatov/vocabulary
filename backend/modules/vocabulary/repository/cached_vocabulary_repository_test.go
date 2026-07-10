package vocabularyrepository

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"vocabulary/backend/modules/vocabulary/service"
)

type mockDbRepo struct {
	calledList   int
	calledCreate int
	mockItems    []vocabularyservice.VocabularyItem
	mockTotal    int
}

func (m *mockDbRepo) Create(ctx context.Context, word, translation, example string, createdBy *string) (*vocabularyservice.VocabularyItem, error) {
	m.calledCreate++
	return &vocabularyservice.VocabularyItem{ID: "id-1", Word: word, Translation: translation}, nil
}

func (m *mockDbRepo) List(ctx context.Context, search string, page, limit int) ([]vocabularyservice.VocabularyItem, int, error) {
	m.calledList++
	return m.mockItems, m.mockTotal, nil
}

func (m *mockDbRepo) AdminList(ctx context.Context, filter vocabularyservice.AdminVocabularySearch) ([]vocabularyservice.VocabularyItem, int, error) {
	return nil, 0, nil
}
func (m *mockDbRepo) AdminCreate(ctx context.Context, input vocabularyservice.AdminVocabularyUpsertInput) (*vocabularyservice.VocabularyItem, error) {
	return nil, nil
}
func (m *mockDbRepo) AdminGet(ctx context.Context, id string) (*vocabularyservice.VocabularyItem, error) {
	return nil, nil
}
func (m *mockDbRepo) AdminUpdate(ctx context.Context, id string, input vocabularyservice.AdminVocabularyUpsertInput) (*vocabularyservice.VocabularyItem, error) {
	return nil, nil
}
func (m *mockDbRepo) AdminDelete(ctx context.Context, id string) error {
	return nil
}
func (m *mockDbRepo) AdminSetStatus(ctx context.Context, id string, status string) (*vocabularyservice.VocabularyItem, error) {
	return nil, nil
}

func TestCachedVocabularyRepository(t *testing.T) {
	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skip("Skipping integration test: Redis is not running locally")
	}
	defer rdb.Close()

	rdb.FlushDB(ctx)

	dbRepo := &mockDbRepo{
		mockItems: []vocabularyservice.VocabularyItem{
			{ID: "1", Word: "test", Translation: "test translation"},
		},
		mockTotal: 1,
	}

	cachedRepo := NewCachedVocabularyRepository(dbRepo, rdb)

	items, total, err := cachedRepo.List(ctx, "test", 1, 10)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if len(items) != 1 || total != 1 {
		t.Errorf("expected 1 item, got len=%d total=%d", len(items), total)
	}
	if dbRepo.calledList != 1 {
		t.Errorf("expected dbList called 1, got %d", dbRepo.calledList)
	}

	time.Sleep(50 * time.Millisecond)

	itemsCached, totalCached, err := cachedRepo.List(ctx, "test", 1, 10)
	if err != nil {
		t.Fatalf("List cached error: %v", err)
	}
	if len(itemsCached) != 1 || totalCached != 1 {
		t.Errorf("expected 1 item from cache, got len=%d total=%d", len(itemsCached), totalCached)
	}
	if dbRepo.calledList != 1 {
		t.Errorf("expected dbList still called 1 (cache hit), got %d", dbRepo.calledList)
	}

	_, err = cachedRepo.Create(ctx, "new", "yangi", "", nil)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	_, _, err = cachedRepo.List(ctx, "test", 1, 10)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if dbRepo.calledList != 2 {
		t.Errorf("expected dbList called 2 (after cache invalidation), got %d", dbRepo.calledList)
	}
}
