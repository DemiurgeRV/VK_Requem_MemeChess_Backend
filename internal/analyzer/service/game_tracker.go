package service

import (
	"fmt"
	"strings"
	"sync"

	"meme_chess/internal/analyzer/analysis"
	"meme_chess/internal/analyzer/position"
)

const (
	defaultTrackedAnalysisDepth = 3
	defaultTrackedFrontierPly   = 2
)

type GameTracker struct {
	analyzer        *AnalyzerService
	mu              sync.RWMutex
	games           map[string][]string
	defaultDepth    int
	defaultFrontier int
}

func NewGameTracker(analyzer *AnalyzerService) *GameTracker {
	return &GameTracker{
		analyzer:        analyzer,
		games:           make(map[string][]string),
		defaultDepth:    defaultTrackedAnalysisDepth,
		defaultFrontier: defaultTrackedFrontierPly,
	}
}

func (t *GameTracker) StartGame(gameID string) {
	gameID = strings.TrimSpace(gameID)
	if gameID == "" {
		return
	}

	t.mu.Lock()
	if _, ok := t.games[gameID]; !ok {
		t.games[gameID] = nil
	}
	t.mu.Unlock()

	if t.analyzer != nil {
		t.analyzer.EnsureHotFrontierAsync(position.NewInitial(), t.defaultDepth, t.defaultFrontier)
	}
}

func (t *GameTracker) RecordMove(gameID, move string) {
	gameID = strings.TrimSpace(gameID)
	move = normalizeUCIMove(move)
	if gameID == "" || move == "" {
		return
	}

	t.mu.Lock()
	stack := append([]string(nil), t.games[gameID]...)
	stack = append(stack, move)
	t.games[gameID] = stack
	t.mu.Unlock()

	if t.analyzer == nil {
		return
	}

	gs, err := position.BuildGameStateFromUCIMoves(stack)
	if err != nil {
		return
	}
	t.analyzer.EnsureHotFrontierAsync(gs, t.defaultDepth, t.defaultFrontier)
}

func (t *GameTracker) AnalyzeRecordedMove(gameID, move string, moveNumber int, depth int) (*analysis.Result, error) {
	if t.analyzer == nil {
		return nil, fmt.Errorf("move analyzer is not configured")
	}

	stack := t.copyStack(gameID)
	if len(stack) == 0 {
		return nil, fmt.Errorf("no tracked moves for game %s", gameID)
	}

	targetIndex, err := selectTrackedMoveIndex(stack, normalizeUCIMove(move), moveNumber)
	if err != nil {
		return nil, err
	}

	before, err := position.BuildGameStateFromUCIMoves(stack[:targetIndex])
	if err != nil {
		return nil, err
	}

	mv, err := position.ParseUCIMove(before, stack[targetIndex])
	if err != nil {
		return nil, err
	}

	if depth < 1 {
		depth = t.defaultDepth
	}

	return t.analyzer.AnalyzeMove(before, mv, depth)
}

func (t *GameTracker) ForgetGame(gameID string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.games, strings.TrimSpace(gameID))
}

func (t *GameTracker) SyncGame(gameID string, moves []string) error {
	normalized := make([]string, 0, len(moves))
	for _, move := range moves {
		move = normalizeUCIMove(move)
		if move == "" {
			continue
		}
		normalized = append(normalized, move)
	}

	if _, err := position.BuildGameStateFromUCIMoves(normalized); err != nil {
		return err
	}

	t.mu.Lock()
	t.games[strings.TrimSpace(gameID)] = normalized
	t.mu.Unlock()
	return nil
}

func (t *GameTracker) copyStack(gameID string) []string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return append([]string(nil), t.games[strings.TrimSpace(gameID)]...)
}

func selectTrackedMoveIndex(stack []string, move string, moveNumber int) (int, error) {
	if moveNumber > 0 {
		index := moveNumber - 1
		if index < 0 || index >= len(stack) {
			return 0, fmt.Errorf("move number %d is outside tracked stack", moveNumber)
		}
		if move != "" && stack[index] != move {
			return 0, fmt.Errorf("tracked move #%d is %s, not %s", moveNumber, stack[index], move)
		}
		return index, nil
	}

	if move == "" {
		return 0, fmt.Errorf("move is required when move_number is not provided")
	}

	for i := len(stack) - 1; i >= 0; i-- {
		if stack[i] == move {
			return i, nil
		}
	}

	return 0, fmt.Errorf("move %s was not found in tracked stack", move)
}

func normalizeUCIMove(move string) string {
	return strings.TrimSpace(strings.ToLower(move))
}
