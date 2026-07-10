package vocabularyrepository

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"vocabulary/backend/modules/vocabulary/service"
)

type CachedVocabularyRepository struct {
	delegate vocabularyservice.VocabularyRepository
	rdb      *redis.Client
}

func NewCachedVocabularyRepository(delegate vocabularyservice.VocabularyRepository, rdb *redis.Client) *CachedVocabularyRepository {
	return &CachedVocabularyRepository{
		delegate: delegate,
		rdb:      rdb,
	}
}

type cachedListResponse struct {
	Items []vocabularyservice.VocabularyItem `json:"items"`
	Total int                                `json:"total"`
}

func (r *CachedVocabularyRepository) List(ctx context.Context, search string, page, limit int) ([]vocabularyservice.VocabularyItem, int, error) {
	key := fmt.Sprintf("vocab:list:search:%s:page:%d:limit:%d", search, page, limit)

	val, err := r.rdb.Get(ctx, key).Result()
	if err == nil {
		var cached cachedListResponse
		if err := json.Unmarshal([]byte(val), &cached); err == nil {
			return cached.Items, cached.Total, nil
		}
	}

	items, total, err := r.delegate.List(ctx, search, page, limit)
	if err != nil {
		return nil, 0, err
	}

	cached := cachedListResponse{Items: items, Total: total}
	if bytes, err := json.Marshal(cached); err == nil {
		_ = r.rdb.Set(ctx, key, bytes, 5*time.Minute).Err()
	}

	return items, total, nil
}

func (r *CachedVocabularyRepository) Create(ctx context.Context, word, translation, example string, createdBy *string) (*vocabularyservice.VocabularyItem, error) {
	item, err := r.delegate.Create(ctx, word, translation, example, createdBy)
	if err != nil {
		return nil, err
	}

	r.invalidateListCaches(ctx)
	return item, nil
}

func (r *CachedVocabularyRepository) invalidateListCaches(ctx context.Context) {
	var cursor uint64
	for {
		keys, nextCursor, err := r.rdb.Scan(ctx, cursor, "vocab:list:*", 100).Result()
		if err != nil {
			log.Printf("redis scan error during cache invalidation: %v", err)
			break
		}
		if len(keys) > 0 {
			_ = r.rdb.Del(ctx, keys...).Err()
		}
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	for {
		keys, nextCursor, err := r.rdb.Scan(ctx, cursor, "vocab:adminlist:*", 100).Result()
		if err != nil {
			log.Printf("redis scan error during admin cache invalidation: %v", err)
			break
		}
		if len(keys) > 0 {
			_ = r.rdb.Del(ctx, keys...).Err()
		}
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
}

func (r *CachedVocabularyRepository) AdminList(ctx context.Context, filter vocabularyservice.AdminVocabularySearch) ([]vocabularyservice.VocabularyItem, int, error) {
	return r.delegate.AdminList(ctx, filter)
}

func (r *CachedVocabularyRepository) AdminCreate(ctx context.Context, input vocabularyservice.AdminVocabularyUpsertInput) (*vocabularyservice.VocabularyItem, error) {
	item, err := r.delegate.AdminCreate(ctx, input)
	if err != nil {
		return nil, err
	}
	r.invalidateListCaches(ctx)
	return item, nil
}

func (r *CachedVocabularyRepository) AdminGet(ctx context.Context, id string) (*vocabularyservice.VocabularyItem, error) {
	key := fmt.Sprintf("vocab:item:%s", id)

	val, err := r.rdb.Get(ctx, key).Result()
	if err == nil {
		var item vocabularyservice.VocabularyItem
		if err := json.Unmarshal([]byte(val), &item); err == nil {
			return &item, nil
		}
	}

	item, err := r.delegate.AdminGet(ctx, id)
	if err != nil {
		return nil, err
	}

	if bytes, err := json.Marshal(item); err == nil {
		_ = r.rdb.Set(ctx, key, bytes, 10*time.Minute).Err()
	}

	return item, nil
}

func (r *CachedVocabularyRepository) AdminUpdate(ctx context.Context, id string, input vocabularyservice.AdminVocabularyUpsertInput) (*vocabularyservice.VocabularyItem, error) {
	item, err := r.delegate.AdminUpdate(ctx, id, input)
	if err != nil {
		return nil, err
	}

	r.invalidateListCaches(ctx)
	_ = r.rdb.Del(ctx, fmt.Sprintf("vocab:item:%s", id)).Err()

	return item, nil
}

func (r *CachedVocabularyRepository) AdminDelete(ctx context.Context, id string) error {
	err := r.delegate.AdminDelete(ctx, id)
	if err != nil {
		return err
	}

	r.invalidateListCaches(ctx)
	_ = r.rdb.Del(ctx, fmt.Sprintf("vocab:item:%s", id)).Err()

	return nil
}

func (r *CachedVocabularyRepository) AdminSetStatus(ctx context.Context, id string, status string) (*vocabularyservice.VocabularyItem, error) {
	item, err := r.delegate.AdminSetStatus(ctx, id, status)
	if err != nil {
		return nil, err
	}

	r.invalidateListCaches(ctx)
	_ = r.rdb.Del(ctx, fmt.Sprintf("vocab:item:%s", id)).Err()

	return item, nil
}
