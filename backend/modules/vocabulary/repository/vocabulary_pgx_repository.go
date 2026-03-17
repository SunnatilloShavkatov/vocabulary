package vocabularyrepository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"vocabulary/backend/modules/vocabulary/service"
)

type VocabularyPgxRepository struct {
	pool *pgxpool.Pool
}

func NewVocabularyPgxRepository(pool *pgxpool.Pool) *VocabularyPgxRepository {
	return &VocabularyPgxRepository{pool: pool}
}

func (r *VocabularyPgxRepository) Create(ctx context.Context, word, translation, example string) (*vocabularyservice.VocabularyItem, error) {
	const q = `
		INSERT INTO vocabularies (id, word, translation, example, created_at)
		VALUES (gen_random_uuid(), $1, $2, NULLIF($3, ''), NOW())
		RETURNING id, word, translation, COALESCE(example, ''), created_at`

	var item vocabularyservice.VocabularyItem
	err := r.pool.QueryRow(ctx, q, word, translation, example).
		Scan(&item.ID, &item.Word, &item.Translation, &item.Example, &item.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *VocabularyPgxRepository) List(ctx context.Context, search string, page, limit int) ([]vocabularyservice.VocabularyItem, int, error) {
	const countQ = `
		SELECT COUNT(*) FROM vocabularies
		WHERE ($1 = '' OR word ILIKE '%' || $1 || '%' OR translation ILIKE '%' || $1 || '%')`

	var total int
	if err := r.pool.QueryRow(ctx, countQ, search).Scan(&total); err != nil {
		return nil, 0, err
	}

	const listQ = `
		SELECT id, word, translation, COALESCE(example, ''), created_at
		FROM vocabularies
		WHERE ($1 = '' OR word ILIKE '%' || $1 || '%' OR translation ILIKE '%' || $1 || '%')
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, listQ, search, limit, (page-1)*limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []vocabularyservice.VocabularyItem
	for rows.Next() {
		var item vocabularyservice.VocabularyItem
		if err := rows.Scan(&item.ID, &item.Word, &item.Translation, &item.Example, &item.CreatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

