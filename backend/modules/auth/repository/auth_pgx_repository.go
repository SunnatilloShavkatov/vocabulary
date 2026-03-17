package authrepository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"vocabulary/backend/modules/auth/service"
)

type AuthPgxRepository struct {
	pool *pgxpool.Pool
}

func NewAuthPgxRepository(pool *pgxpool.Pool) *AuthPgxRepository {
	return &AuthPgxRepository{pool: pool}
}

func (r *AuthPgxRepository) CreateAdmin(ctx context.Context, email, passwordHash, role string) (*authservice.AuthAdmin, error) {
	const q = `
		INSERT INTO admins (id, email, password_hash, role, created_at)
		VALUES (gen_random_uuid(), $1, $2, $3, NOW())
		RETURNING id, email, role, created_at`

	var admin authservice.AuthAdmin
	err := r.pool.QueryRow(ctx, q, email, passwordHash, role).Scan(&admin.ID, &admin.Email, &admin.Role, &admin.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, authservice.ErrAuthAdminAlreadyExists
		}
		return nil, err
	}

	return &admin, nil
}

func (r *AuthPgxRepository) FindAdminByEmail(ctx context.Context, email string) (*authservice.AuthAdminCredentials, error) {
	const q = `
		SELECT id, email, password_hash, role, created_at
		FROM admins
		WHERE email = $1`

	var admin authservice.AuthAdminCredentials
	err := r.pool.QueryRow(ctx, q, email).Scan(&admin.ID, &admin.Email, &admin.PasswordHash, &admin.Role, &admin.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, authservice.ErrAuthAdminNotFound
		}
		return nil, err
	}

	return &admin, nil
}

