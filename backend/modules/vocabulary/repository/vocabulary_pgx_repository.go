package vocabularyrepository

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"vocabulary/backend/modules/vocabulary/service"
)

type VocabularyPgxRepository struct {
	pool *pgxpool.Pool
}

func NewVocabularyPgxRepository(pool *pgxpool.Pool) *VocabularyPgxRepository {
	return &VocabularyPgxRepository{pool: pool}
}

func (r *VocabularyPgxRepository) Create(ctx context.Context, word, translation, example string, createdBy *string) (*vocabularyservice.VocabularyItem, error) {
	const q = `
		INSERT INTO vocabularies (id, word, translation, example, category, status, created_by, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, NULLIF($3, ''), NULL, 'pending', $4, NOW(), NOW())
		RETURNING id, word, translation, COALESCE(example, ''), COALESCE(category, ''), COALESCE(status, 'pending'), COALESCE(updated_at, created_at), created_at`

	var item vocabularyservice.VocabularyItem
	err := r.pool.QueryRow(ctx, q, word, translation, example, createdBy).
		Scan(&item.ID, &item.Word, &item.Translation, &item.Example, &item.Category, &item.Status, &item.UpdatedAt, &item.CreatedAt)
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
		SELECT id, word, translation, COALESCE(example, ''), COALESCE(category, ''), COALESCE(status, 'pending'), COALESCE(updated_at, created_at), created_at
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
		if err := rows.Scan(&item.ID, &item.Word, &item.Translation, &item.Example, &item.Category, &item.Status, &item.UpdatedAt, &item.CreatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r *VocabularyPgxRepository) AdminList(ctx context.Context, filter vocabularyservice.AdminVocabularySearch) ([]vocabularyservice.VocabularyItem, int, error) {
	const countQ = `
		SELECT COUNT(*)
		FROM vocabularies
		WHERE ($1 = '' OR word ILIKE '%' || $1 || '%')
		  AND ($2 = '' OR translation ILIKE '%' || $2 || '%')
		  AND ($3 = '' OR COALESCE(category, '') ILIKE '%' || $3 || '%')
		  AND ($4 = '' OR status = $4)`

	var total int
	if err := r.pool.QueryRow(ctx, countQ, filter.Word, filter.Translation, filter.Category, filter.Status).Scan(&total); err != nil {
		return nil, 0, err
	}

	const listQ = `
		SELECT id, word, translation, COALESCE(example, ''), COALESCE(category, ''), COALESCE(status, 'pending'), COALESCE(updated_at, created_at), created_at
		FROM vocabularies
		WHERE ($1 = '' OR word ILIKE '%' || $1 || '%')
		  AND ($2 = '' OR translation ILIKE '%' || $2 || '%')
		  AND ($3 = '' OR COALESCE(category, '') ILIKE '%' || $3 || '%')
		  AND ($4 = '' OR status = $4)
		ORDER BY created_at DESC
		LIMIT $5 OFFSET $6`

	rows, err := r.pool.Query(ctx, listQ, filter.Word, filter.Translation, filter.Category, filter.Status, filter.Limit, (filter.Page-1)*filter.Limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]vocabularyservice.VocabularyItem, 0)
	for rows.Next() {
		var item vocabularyservice.VocabularyItem
		if err := rows.Scan(&item.ID, &item.Word, &item.Translation, &item.Example, &item.Category, &item.Status, &item.UpdatedAt, &item.CreatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}

	return items, total, rows.Err()
}

func (r *VocabularyPgxRepository) AdminCreate(ctx context.Context, input vocabularyservice.AdminVocabularyUpsertInput) (*vocabularyservice.VocabularyItem, error) {
	const q = `
		INSERT INTO vocabularies (id, word, translation, example, category, status, created_by, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, NULLIF($3, ''), NULLIF($4, ''), $5, NULLIF($6, '')::uuid, NOW(), NOW())
		RETURNING id, word, translation, COALESCE(example, ''), COALESCE(category, ''), COALESCE(status, 'pending'), COALESCE(updated_at, created_at), created_at`

	var item vocabularyservice.VocabularyItem
	err := r.pool.QueryRow(ctx, q, input.Word, input.Translation, input.Example, input.Category, input.Status, nullableUUIDText(input.CreatedBy)).
		Scan(&item.ID, &item.Word, &item.Translation, &item.Example, &item.Category, &item.Status, &item.UpdatedAt, &item.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return nil, vocabularyservice.ErrInvalidVocabularyInput
		}
		return nil, err
	}
	return &item, nil
}

func (r *VocabularyPgxRepository) AdminGet(ctx context.Context, id string) (*vocabularyservice.VocabularyItem, error) {
	const q = `
		SELECT id, word, translation, COALESCE(example, ''), COALESCE(category, ''), COALESCE(status, 'pending'), COALESCE(updated_at, created_at), created_at
		FROM vocabularies
		WHERE id::text = $1`

	var item vocabularyservice.VocabularyItem
	err := r.pool.QueryRow(ctx, q, strings.TrimSpace(id)).
		Scan(&item.ID, &item.Word, &item.Translation, &item.Example, &item.Category, &item.Status, &item.UpdatedAt, &item.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, vocabularyservice.ErrVocabularyNotFound
		}
		return nil, err
	}
	return &item, nil
}

func (r *VocabularyPgxRepository) AdminUpdate(ctx context.Context, id string, input vocabularyservice.AdminVocabularyUpsertInput) (*vocabularyservice.VocabularyItem, error) {
	const q = `
		UPDATE vocabularies
		SET word = $2,
		    translation = $3,
		    example = NULLIF($4, ''),
		    category = NULLIF($5, ''),
		    status = COALESCE(NULLIF($6, ''), status),
		    updated_at = NOW()
		WHERE id::text = $1
		RETURNING id, word, translation, COALESCE(example, ''), COALESCE(category, ''), COALESCE(status, 'pending'), COALESCE(updated_at, created_at), created_at`

	var item vocabularyservice.VocabularyItem
	err := r.pool.QueryRow(ctx, q, strings.TrimSpace(id), input.Word, input.Translation, input.Example, input.Category, input.Status).
		Scan(&item.ID, &item.Word, &item.Translation, &item.Example, &item.Category, &item.Status, &item.UpdatedAt, &item.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, vocabularyservice.ErrVocabularyNotFound
		}
		return nil, err
	}
	return &item, nil
}

func (r *VocabularyPgxRepository) AdminDelete(ctx context.Context, id string) error {
	const q = `DELETE FROM vocabularies WHERE id::text = $1`
	res, err := r.pool.Exec(ctx, q, strings.TrimSpace(id))
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return vocabularyservice.ErrVocabularyNotFound
	}
	return nil
}

func (r *VocabularyPgxRepository) AdminSetStatus(ctx context.Context, id string, status string) (*vocabularyservice.VocabularyItem, error) {
	const q = `
		UPDATE vocabularies
		SET status = $2, updated_at = NOW()
		WHERE id::text = $1
		RETURNING id, word, translation, COALESCE(example, ''), COALESCE(category, ''), COALESCE(status, 'pending'), COALESCE(updated_at, created_at), created_at`

	var item vocabularyservice.VocabularyItem
	err := r.pool.QueryRow(ctx, q, strings.TrimSpace(id), strings.TrimSpace(status)).
		Scan(&item.ID, &item.Word, &item.Translation, &item.Example, &item.Category, &item.Status, &item.UpdatedAt, &item.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, vocabularyservice.ErrVocabularyNotFound
		}
		return nil, err
	}
	return &item, nil
}

func nullableUUIDText(v string) any {
	v = strings.TrimSpace(v)
	if v == "" {
		return nil
	}
	return v
}

