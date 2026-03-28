package game

import "sync"

type Status string

const (
	StatusWaiting  Status = "waiting"
	StatusActive   Status = "active"
	StatusFinished Status = "finished"
)

type Move struct {
	Number int    `json:"number"`
	UserID string `json:"user_id"`
	Move   string `json:"move"`
}

type Session struct {
	mu sync.RWMutex

	GameID string `json:"game_id"`

	Player1ID string `json:"player1_id"`
	Player2ID string `json:"player2_id"`

	Player1Connected bool `json:"player1_connected"`
	Player2Connected bool `json:"player2_connected"`

	Status Status `json:"status"`

	CurrentTurnUserID string `json:"current_turn_user_id"`

	Moves []Move `json:"moves"`
}

func NewSession(gameID, player1ID, player2ID string) *Session {
	return &Session{
		GameID:            gameID,
		Player1ID:         player1ID,
		Player2ID:         player2ID,
		Status:            StatusWaiting,
		CurrentTurnUserID: player1ID,
		Moves:             make([]Move, 0, 64),
		Player1Connected:  false,
		Player2Connected:  false,
	}
}

func (s *Session) HasPlayer(userID string) bool {
	return userID == s.Player1ID || userID == s.Player2ID
}

func (s *Session) SetConnected(userID string, connected bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if userID == s.Player1ID {
		s.Player1Connected = connected
	}
	if userID == s.Player2ID {
		s.Player2Connected = connected
	}

	if s.Player1Connected && s.Player2Connected && s.Status == StatusWaiting {
		s.Status = StatusActive
	}
}

func (s *Session) Snapshot() State {
	s.mu.RLock()
	defer s.mu.RUnlock()

	moves := make([]Move, len(s.Moves))
	copy(moves, s.Moves)

	return State{
		GameID:            s.GameID,
		Player1ID:         s.Player1ID,
		Player2ID:         s.Player2ID,
		Player1Connected:  s.Player1Connected,
		Player2Connected:  s.Player2Connected,
		Status:            string(s.Status),
		CurrentTurnUserID: s.CurrentTurnUserID,
		Moves:             moves,
	}
}

func (s *Session) ApplyMove(userID, move string) (State, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.Status == StatusFinished {
		return State{}, ErrGameFinished
	}
	if s.Status != StatusActive {
		return State{}, ErrGameNotActive
	}
	if userID != s.CurrentTurnUserID {
		return State{}, ErrNotYourTurn
	}

	nextMove := Move{
		Number: len(s.Moves) + 1,
		UserID: userID,
		Move:   move,
	}
	s.Moves = append(s.Moves, nextMove)

	if s.CurrentTurnUserID == s.Player1ID {
		s.CurrentTurnUserID = s.Player2ID
	} else {
		s.CurrentTurnUserID = s.Player1ID
	}

	moves := make([]Move, len(s.Moves))
	copy(moves, s.Moves)

	return State{
		GameID:            s.GameID,
		Player1ID:         s.Player1ID,
		Player2ID:         s.Player2ID,
		Player1Connected:  s.Player1Connected,
		Player2Connected:  s.Player2Connected,
		Status:            string(s.Status),
		CurrentTurnUserID: s.CurrentTurnUserID,
		Moves:             moves,
	}, nil
}
