package game

import (
	"fmt"
	"strings"

	"meme_chess/internal/analyzer/analysis"
)

func (s *Service) AnalyzeMove(gameID, userID, move string, moveNumber int, depth int) (*analysis.Result, error) {
	session, ok := s.GetSession(gameID)
	if !ok {
		return nil, ErrGameNotFound
	}
	if !session.HasPlayer(userID) {
		return nil, ErrForbidden
	}
	if s.moveAnalyzer == nil {
		return nil, fmt.Errorf("move analyzer is not configured")
	}

	move = strings.TrimSpace(strings.ToLower(move))
	if move == "" && moveNumber < 1 {
		return nil, fmt.Errorf("move is required")
	}

	if err := s.syncAnalyzerStack(gameID, session); err != nil {
		return nil, err
	}

	return s.moveAnalyzer.AnalyzeRecordedMove(gameID, move, moveNumber, depth)
}

func (s *Service) syncAnalyzerStack(gameID string, session *Session) error {
	snapshot := session.Snapshot()
	stack := make([]string, 0, len(snapshot.Moves))
	for _, move := range snapshot.Moves {
		stack = append(stack, move.Move)
	}

	type syncer interface {
		SyncGame(gameID string, moves []string) error
	}

	moveAnalyzer, ok := s.moveAnalyzer.(syncer)
	if !ok {
		return nil
	}

	return moveAnalyzer.SyncGame(gameID, stack)
}
