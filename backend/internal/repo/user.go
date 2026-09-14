package repo

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kxs888/kxs/backend/internal/domain"
	"github.com/kxs888/kxs/backend/internal/errcode"
)

type UserRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

func (r *UserRepo) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, username, password_hash, display_name, created_at
		FROM users WHERE username = $1`, username)
	return scanUser(row)
}

func (r *UserRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, username, password_hash, display_name, created_at
		FROM users WHERE id = $1`, id)
	return scanUser(row)
}

func (r *UserRepo) Insert(ctx context.Context, u *domain.User) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO users (username, password_hash, display_name)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`,
		u.Username, u.PasswordHash, u.DisplayName,
	).Scan(&u.ID, &u.CreatedAt)
}

func (r *UserRepo) Count(ctx context.Context) (int64, error) {
	var n int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&n)
	return n, err
}

func scanUser(row pgx.Row) (*domain.User, error) {
	var u domain.User
	err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.DisplayName, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errcode.NotFound("user not found")
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}
