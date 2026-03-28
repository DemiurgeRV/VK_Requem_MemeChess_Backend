package game

import (
	"errors"
	"sync"
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
	Moves             []Move `json:"moves"`
}

type Service struct {
	mu       sync.RWMutex
	sessions map[string]*Session
}

func NewService() *Service {
	return &Service{
		sessions: make(map[string]*Session),
	}
}

func (s *Service) CreateGame(gameID, player1ID, player2ID string) *Session {
	s.mu.Lock()
	defer s.mu.Unlock()

	session := NewSession(gameID, player1ID, player2ID)
	s.sessions[gameID] = session
	return session
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

func (s *Service) MakeMove(gameID, userID, move string) (State, error) {
	session, ok := s.GetSession(gameID)
	if !ok {
		return State{}, ErrGameNotFound
	}
	if !session.HasPlayer(userID) {
		return State{}, ErrForbidden
	}
	if move == "" {
		return State{}, ErrInvalidMove
	}

	return session.ApplyMove(userID, move)
}
