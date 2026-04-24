package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

type User struct {
	ID           string
	Email        *string
	Username     string
	AvatarURL    *string
	CreatedAt    time.Time
	PasswordHash string
}

func (r *Repository) Create(ctx context.Context, username string, email *string, passwordHash string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	const q = `
		INSERT INTO users (id, username, email, password_hash)
		VALUES (gen_random_uuid(), $1, $2, $3)
		RETURNING id::text
	`

	var id string
	err := r.pool.QueryRow(ctx, q, username, email, passwordHash).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("insert user: %w", err)
	}
	return id, nil
}

func (r *Repository) GetByLogin(ctx context.Context, login string) (*User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	const q = `
		SELECT id::text, email, username, avatar_url, created_at, password_hash
		FROM users
		WHERE lower(username) = lower($1)
		   OR (email IS NOT NULL AND lower(email) = lower($1))
	`

	var u User
	err := r.pool.QueryRow(ctx, q, login).Scan(
		&u.ID,
		&u.Email,
		&u.Username,
		&u.AvatarURL,
		&u.CreatedAt,
		&u.PasswordHash,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("select user: %w", err)
	}
	return &u, nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (*User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	const q = `
		SELECT id::text, email, username, avatar_url, created_at, password_hash
		FROM users
		WHERE id = $1
	`

	var u User
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&u.ID,
		&u.Email,
		&u.Username,
		&u.AvatarURL,
		&u.CreatedAt,
		&u.PasswordHash,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("select user: %w", err)
	}
	return &u, nil
}
