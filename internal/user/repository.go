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
	ShopCurrency int64
	GameCurrency int64
	CreatedAt    time.Time
	PasswordHash string
}

func (r *Repository) Create(ctx context.Context, username string, email *string, passwordHash string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	const q = `
		INSERT INTO users (id, username, email, password_hash, shop_currency, game_currency)
		VALUES (gen_random_uuid(), $1, $2, $3, 0, 1000)
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
		SELECT id::text, email, username, avatar_url, shop_currency, game_currency, created_at, password_hash
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
		&u.ShopCurrency,
		&u.GameCurrency,
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
		SELECT id::text, email, username, avatar_url, shop_currency, game_currency, created_at, password_hash
		FROM users
		WHERE id = $1
	`

	var u User
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&u.ID,
		&u.Email,
		&u.Username,
		&u.AvatarURL,
		&u.ShopCurrency,
		&u.GameCurrency,
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

var ErrInsufficientGameCurrency = errors.New("insufficient game currency")

func (r *Repository) ReserveGameCurrency(ctx context.Context, userID string, amount int64) error {
	if amount <= 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	const q = `
		UPDATE users
		SET game_currency = game_currency - $2
		WHERE id = $1
		  AND game_currency >= $2
	`

	tag, err := r.pool.Exec(ctx, q, userID, amount)
	if err != nil {
		return fmt.Errorf("reserve game currency: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrInsufficientGameCurrency
	}
	return nil
}

func (r *Repository) AddGameCurrency(ctx context.Context, userID string, amount int64) error {
	if amount <= 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	const q = `
		UPDATE users
		SET game_currency = game_currency + $2
		WHERE id = $1
	`

	_, err := r.pool.Exec(ctx, q, userID, amount)
	if err != nil {
		return fmt.Errorf("add game currency: %w", err)
	}
	return nil
}
