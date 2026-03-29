package game

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

type CreateGameParams struct {
	GameID            string
	Player1ID         string
	Player2ID         string
	Status            string
	FEN               string
	CurrentTurnUserID string
}

func (r *Repository) CreateGame(ctx context.Context, p CreateGameParams) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `
		INSERT INTO games (
			id, player1_id, player2_id, status, fen, current_turn_user_id
		) VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.pool.Exec(ctx, query,
		p.GameID,
		p.Player1ID,
		p.Player2ID,
		p.Status,
		p.FEN,
		p.CurrentTurnUserID,
	)
	if err != nil {
		return fmt.Errorf("insert game: %w", err)
	}

	return nil
}

type SaveMoveParams struct {
	GameID      string
	PlayerID    string
	MoveNumber  int
	Move        string
	FEN         string
	IsCapture   bool
	IsCheck     bool
	IsCheckmate bool
}

func (r *Repository) SaveMove(ctx context.Context, p SaveMoveParams) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `
		INSERT INTO moves (
			game_id, player_id, move_number, move, fen, is_capture, is_check, is_checkmate
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.pool.Exec(ctx, query,
		p.GameID,
		p.PlayerID,
		p.MoveNumber,
		p.Move,
		p.FEN,
		p.IsCapture,
		p.IsCheck,
		p.IsCheckmate,
	)
	if err != nil {
		return fmt.Errorf("insert move: %w", err)
	}

	return nil
}

type UpdateGameStateParams struct {
	GameID            string
	Status            string
	FEN               string
	CurrentTurnUserID string
	WinnerID          *string
	FinishedAt        *time.Time
}

func (r *Repository) UpdateGameState(ctx context.Context, p UpdateGameStateParams) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `
		UPDATE games
		SET
			status = $2,
			fen = $3,
			current_turn_user_id = $4,
			winner_id = $5,
			finished_at = $6
		WHERE id = $1
	`

	_, err := r.pool.Exec(ctx, query,
		p.GameID,
		p.Status,
		p.FEN,
		p.CurrentTurnUserID,
		p.WinnerID,
		p.FinishedAt,
	)
	if err != nil {
		return fmt.Errorf("update game state: %w", err)
	}

	return nil
}
