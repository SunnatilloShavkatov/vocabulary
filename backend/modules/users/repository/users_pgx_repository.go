package usersrepository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	usersservice "vocabulary/backend/modules/users/service"
)

type UsersPgxRepository struct {
	pool *pgxpool.Pool
}

func NewUsersPgxRepository(pool *pgxpool.Pool) *UsersPgxRepository {
	return &UsersPgxRepository{pool: pool}
}

func (r *UsersPgxRepository) FindProfile(ctx context.Context, userID string) (*usersservice.UserProfile, error) {
	const q = `
		SELECT id, role, settings
		FROM users
		WHERE id::text = $1`

	var (
		p            usersservice.UserProfile
		settingsJSON []byte
	)
	if err := r.pool.QueryRow(ctx, q, userID).Scan(&p.ID, &p.Role, &settingsJSON); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if len(settingsJSON) == 0 {
		p.Settings = map[string]string{}
		return &p, nil
	}
	if err := json.Unmarshal(settingsJSON, &p.Settings); err != nil {
		p.Settings = map[string]string{}
	}
	return &p, nil
}

func (r *UsersPgxRepository) AdminList(ctx context.Context, query string) ([]usersservice.AdminUser, error) {
	const q = `
		SELECT id, COALESCE(name, ''), email, role, COALESCE(status, 'active'), created_at, COALESCE(updated_at, created_at)
		FROM users
		WHERE ($1 = '' OR COALESCE(name, '') ILIKE '%' || $1 || '%' OR email ILIKE '%' || $1 || '%')
		ORDER BY created_at ASC`

	rows, err := r.pool.Query(ctx, q, strings.TrimSpace(query))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]usersservice.AdminUser, 0)
	for rows.Next() {
		var u usersservice.AdminUser
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Role, &u.Status, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, u)
	}
	return items, rows.Err()
}

func (r *UsersPgxRepository) AdminCreate(ctx context.Context, input usersservice.AdminCreateUserInput) (*usersservice.AdminUser, error) {
	name := strings.TrimSpace(input.Name)
	email := strings.TrimSpace(strings.ToLower(input.Email))
	if name == "" || !isValidEmail(email) {
		return nil, usersservice.ErrInvalidAdminUserInput
	}

	const q = `
		INSERT INTO users (id, name, email, role, status, settings, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, 'client', 'active', '{}'::jsonb, NOW(), NOW())
		RETURNING id, name, email, role, status, created_at, updated_at`

	var u usersservice.AdminUser
	if err := r.pool.QueryRow(ctx, q, name, email).Scan(&u.ID, &u.Name, &u.Email, &u.Role, &u.Status, &u.CreatedAt, &u.UpdatedAt); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, usersservice.ErrAdminUserAlreadyExists
		}
		return nil, err
	}
	return &u, nil
}

func (r *UsersPgxRepository) AdminGet(ctx context.Context, id string) (*usersservice.AdminUser, error) {
	const q = `
		SELECT id, COALESCE(name, ''), email, role, COALESCE(status, 'active'), created_at, COALESCE(updated_at, created_at)
		FROM users
		WHERE id::text = $1`

	var u usersservice.AdminUser
	if err := r.pool.QueryRow(ctx, q, strings.TrimSpace(id)).Scan(&u.ID, &u.Name, &u.Email, &u.Role, &u.Status, &u.CreatedAt, &u.UpdatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, usersservice.ErrAdminUserNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *UsersPgxRepository) AdminUpdate(ctx context.Context, id string, input usersservice.AdminUpdateUserInput) (*usersservice.AdminUser, error) {
	name := strings.TrimSpace(input.Name)
	email := strings.TrimSpace(strings.ToLower(input.Email))
	if name == "" || !isValidEmail(email) {
		return nil, usersservice.ErrInvalidAdminUserInput
	}

	const q = `
		UPDATE users
		SET name = $2, email = $3, updated_at = NOW()
		WHERE id::text = $1
		RETURNING id, name, email, role, COALESCE(status, 'active'), created_at, COALESCE(updated_at, created_at)`

	var u usersservice.AdminUser
	if err := r.pool.QueryRow(ctx, q, strings.TrimSpace(id), name, email).Scan(&u.ID, &u.Name, &u.Email, &u.Role, &u.Status, &u.CreatedAt, &u.UpdatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, usersservice.ErrAdminUserNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, usersservice.ErrAdminUserAlreadyExists
		}
		return nil, err
	}
	return &u, nil
}

func (r *UsersPgxRepository) AdminDelete(ctx context.Context, id string) error {
	const q = `DELETE FROM users WHERE id::text = $1`
	res, err := r.pool.Exec(ctx, q, strings.TrimSpace(id))
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return usersservice.ErrAdminUserNotFound
	}
	return nil
}

func (r *UsersPgxRepository) AdminSetStatus(ctx context.Context, id string, status string) (*usersservice.AdminUser, error) {
	const q = `
		UPDATE users
		SET status = $2, updated_at = NOW()
		WHERE id::text = $1
		RETURNING id, COALESCE(name, ''), email, role, COALESCE(status, 'active'), created_at, COALESCE(updated_at, created_at)`

	var u usersservice.AdminUser
	if err := r.pool.QueryRow(ctx, q, strings.TrimSpace(id), strings.TrimSpace(status)).Scan(&u.ID, &u.Name, &u.Email, &u.Role, &u.Status, &u.CreatedAt, &u.UpdatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, usersservice.ErrAdminUserNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *UsersPgxRepository) AdminStats(ctx context.Context) (*usersservice.AdminStats, error) {
	stats := &usersservice.AdminStats{}
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&stats.UsersCount); err != nil {
		return nil, err
	}
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM vocabularies`).Scan(&stats.WordsCount); err != nil {
		return nil, err
	}
	return stats, nil
}

func (r *UsersPgxRepository) AdminExport(ctx context.Context, format string) (*usersservice.AdminExportResult, error) {
	items, err := r.AdminList(ctx, "")
	if err != nil {
		return nil, err
	}
	stats, err := r.AdminStats(ctx)
	if err != nil {
		return nil, err
	}

	f := strings.ToLower(strings.TrimSpace(format))
	if f == "" {
		f = "json"
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.Before(items[j].CreatedAt)
	})

	switch f {
	case "json":
		payload, err := json.MarshalIndent(map[string]any{
			"exported_at": time.Now().UTC().Format(time.RFC3339),
			"stats":       stats,
			"users":       items,
		}, "", "  ")
		if err != nil {
			return nil, err
		}
		return &usersservice.AdminExportResult{Format: "json", ContentType: "application/json", Filename: "admin_export.json", Data: payload}, nil
	case "sql":
		var b strings.Builder
		b.WriteString("-- admin export users\n")
		for _, u := range items {
			b.WriteString(fmt.Sprintf("INSERT INTO users (id, name, email, role, status) VALUES ('%s','%s','%s','%s','%s');\n",
				escapeSQL(u.ID), escapeSQL(u.Name), escapeSQL(u.Email), escapeSQL(u.Role), escapeSQL(u.Status)))
		}
		return &usersservice.AdminExportResult{Format: "sql", ContentType: "application/sql", Filename: "admin_export.sql", Data: []byte(b.String())}, nil
	default:
		return nil, usersservice.ErrUnsupportedExportFormat
	}
}

func (r *UsersPgxRepository) CreateAuditLog(ctx context.Context, input usersservice.AuditLogCreateInput) (*usersservice.AuditLog, error) {
	input.ActorID = strings.TrimSpace(input.ActorID)
	input.Action = strings.TrimSpace(strings.ToLower(input.Action))
	input.TargetType = strings.TrimSpace(strings.ToLower(input.TargetType))
	input.TargetID = strings.TrimSpace(input.TargetID)
	if input.ActorID == "" || input.Action == "" || input.TargetType == "" {
		return nil, usersservice.ErrInvalidAuditLogInput
	}

	metadata := input.Metadata
	if metadata == nil {
		metadata = map[string]string{}
	}
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return nil, err
	}

	const q = `
		INSERT INTO audit_logs (id, actor_id, action, target_type, target_id, metadata, created_at)
		VALUES (gen_random_uuid(), $1, $2, $3, NULLIF($4, ''), $5::jsonb, NOW())
		RETURNING id, actor_id, action, target_type, COALESCE(target_id, ''), metadata, created_at`

	var (
		entry        usersservice.AuditLog
		metadataData []byte
	)
	if err := r.pool.QueryRow(ctx, q, input.ActorID, input.Action, input.TargetType, input.TargetID, metadataJSON).
		Scan(&entry.ID, &entry.ActorID, &entry.Action, &entry.TargetType, &entry.TargetID, &metadataData, &entry.CreatedAt); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return nil, usersservice.ErrInvalidAuditLogInput
		}
		return nil, err
	}

	entry.Metadata = map[string]string{}
	if len(metadataData) > 0 {
		_ = json.Unmarshal(metadataData, &entry.Metadata)
	}
	return &entry, nil
}

func (r *UsersPgxRepository) ListAuditLogs(ctx context.Context, filter usersservice.AuditLogListFilter) ([]usersservice.AuditLog, int, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 || filter.Limit > 100 {
		filter.Limit = 20
	}
	filter.ActorID = strings.TrimSpace(filter.ActorID)
	filter.Action = strings.TrimSpace(strings.ToLower(filter.Action))

	const countQ = `
		SELECT COUNT(*)
		FROM audit_logs
		WHERE ($1 = '' OR actor_id::text = $1)
		  AND ($2 = '' OR action = $2)`

	var total int
	if err := r.pool.QueryRow(ctx, countQ, filter.ActorID, filter.Action).Scan(&total); err != nil {
		return nil, 0, err
	}

	const listQ = `
		SELECT id, actor_id, action, target_type, COALESCE(target_id, ''), metadata, created_at
		FROM audit_logs
		WHERE ($1 = '' OR actor_id::text = $1)
		  AND ($2 = '' OR action = $2)
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4`

	rows, err := r.pool.Query(ctx, listQ, filter.ActorID, filter.Action, filter.Limit, (filter.Page-1)*filter.Limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]usersservice.AuditLog, 0)
	for rows.Next() {
		var (
			item         usersservice.AuditLog
			metadataData []byte
		)
		if err := rows.Scan(&item.ID, &item.ActorID, &item.Action, &item.TargetType, &item.TargetID, &metadataData, &item.CreatedAt); err != nil {
			return nil, 0, err
		}
		item.Metadata = map[string]string{}
		if len(metadataData) > 0 {
			_ = json.Unmarshal(metadataData, &item.Metadata)
		}
		items = append(items, item)
	}

	return items, total, rows.Err()
}

func isValidEmail(email string) bool {
	email = strings.TrimSpace(email)
	if email == "" {
		return false
	}
	at := strings.Index(email, "@")
	return at > 0 && at < len(email)-1
}

func escapeSQL(v string) string {
	return strings.ReplaceAll(v, "'", "''")
}
