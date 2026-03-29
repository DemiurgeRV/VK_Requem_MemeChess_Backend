package game

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	ErrGameNotFound  = errors.New("game not found")
	ErrForbidden     = errors.New("forbidden")
	ErrGameFull      = errors.New("game room is full")
	ErrNotYourTurn   = errors.New("not your turn")
	ErrGameFinished  = errors.New("game already finished")
	ErrGameNotActive = errors.New("game is not active")
	ErrInvalidMove   = errors.New("invalid move")
)

type State struct {
	GameID            string `json:"game_id"`
	Player1ID         string `json:"player1_id"`
	Player2ID         string `json:"player2_id"`
	Player1Connected  bool   `json:"player1_connected"`
	Player2Connected  bool   `json:"player2_connected"`
	Status            string `json:"status"`
	CurrentTurnUserID string `json:"current_turn_user_id"`
	FEN               string `json:"fen"`
	LastMove          string `json:"last_move"`
	WinnerID          string `json:"winner_id,omitempty"`
	FinishedReason    string `json:"finished_reason,omitempty"`
	Moves             []Move `json:"moves"`
}

type Service struct {
	mu         sync.RWMutex
	sessions   map[string]*Session
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{
		sessions:   make(map[string]*Session),
		repository: repository,
	}
}

func (s *Service) CreateGame(ctx context.Context, gameID, player1ID, player2ID string, engine Engine) (*Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	session := NewSession(gameID, player1ID, player2ID, engine)
	s.sessions[gameID] = session

	if s.repository != nil {
		err := s.repository.CreateGame(ctx, CreateGameParams{
			GameID:            gameID,
			Player1ID:         player1ID,
			Player2ID:         player2ID,
			Status:            string(session.Status),
			FEN:               session.FEN,
			CurrentTurnUserID: session.CurrentTurnUserID,
		})
		if err != nil {
			delete(s.sessions, gameID)
			return nil, err
		}
	}

	return session, nil
}

func (s *Service) GetSession(gameID string) (*Session, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, ok := s.sessions[gameID]
	return session, ok
}

func (s *Service) JoinGame(gameID, userID string) (State, error) {
	session, ok := s.GetSession(gameID)
	if !ok {
		return State{}, ErrGameNotFound
	}

	if !session.HasPlayer(userID) {
		return State{}, ErrForbidden
	}

	session.SetConnected(userID, true)
	return session.Snapshot(), nil
}

func (s *Service) LeaveGame(gameID, userID string) error {
	session, ok := s.GetSession(gameID)
	if !ok {
		return ErrGameNotFound
	}
	if !session.HasPlayer(userID) {
		return ErrForbidden
	}

	session.SetConnected(userID, false)
	return nil
}

func (s *Service) MakeMove(ctx context.Context, gameID, userID, move string) (State, MoveResult, error) {
	session, ok := s.GetSession(gameID)
	if !ok {
		return State{}, MoveResult{}, ErrGameNotFound
	}
	if !session.HasPlayer(userID) {
		return State{}, MoveResult{}, ErrForbidden
	}
	if move == "" {
		return State{}, MoveResult{}, ErrInvalidMove
	}

	state, result, err := session.ApplyMove(userID, move)
	if err != nil {
		return State{}, MoveResult{}, err
	}

	if s.repository != nil {
		moveNumber := len(state.Moves)

		if err := s.repository.SaveMove(ctx, SaveMoveParams{
			GameID:      gameID,
			PlayerID:    userID,
			MoveNumber:  moveNumber,
			Move:        result.Move,
			FEN:         result.FEN,
			IsCapture:   result.IsCapture,
			IsCheck:     result.IsCheck,
			IsCheckmate: result.IsCheckmate,
		}); err != nil {
			return State{}, MoveResult{}, err
		}

		var winnerID *string
		var finishedAt *time.Time

		if state.WinnerID != "" {
			winnerID = &state.WinnerID
			now := time.Now()
			finishedAt = &now
		}

		if err := s.repository.UpdateGameState(ctx, UpdateGameStateParams{
			GameID:            gameID,
			Status:            state.Status,
			FEN:               state.FEN,
			CurrentTurnUserID: state.CurrentTurnUserID,
			WinnerID:          winnerID,
			FinishedAt:        finishedAt,
		}); err != nil {
			return State{}, MoveResult{}, err
		}
	}

	return state, result, nil
}
