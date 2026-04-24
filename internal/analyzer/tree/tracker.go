package tree

import (
	"errors"
	"strings"
	"sync"

	"meme_chess/internal/analyzer/position"
)

var ErrGameNotTracked = errors.New("game is not tracked in variant tree")

type CursorSnapshot struct {
	GameID              string `json:"game_id"`
	RootPositionHash    string `json:"root_position_hash"`
	CurrentPositionHash string `json:"current_position_hash"`
	Ply                 int    `json:"ply"`
}

type PositionNode struct {
	PositionHash string            `json:"position_hash"`
	FEN          string            `json:"fen"`
	Children     map[string]string `json:"children"`
}

type Tracker struct {
	mu    sync.RWMutex
	nodes map[string]*PositionNode
	games map[string]*gameCursor
}

type gameCursor struct {
	rootHash    string
	currentHash string
	currentFEN  string
	ply         int
}

func NewTracker() *Tracker {
	return &Tracker{
		nodes: make(map[string]*PositionNode),
		games: make(map[string]*gameCursor),
	}
}

func (t *Tracker) TrackGame(gameID, initialFEN string) CursorSnapshot {
	t.mu.Lock()
	defer t.mu.Unlock()

	rootHash := position.HashFEN(initialFEN)
	t.ensureNodeLocked(rootHash, initialFEN)
	t.games[gameID] = &gameCursor{
		rootHash:    rootHash,
		currentHash: rootHash,
		currentFEN:  initialFEN,
	}

	return CursorSnapshot{
		GameID:              gameID,
		RootPositionHash:    rootHash,
		CurrentPositionHash: rootHash,
		Ply:                 0,
	}
}

func (t *Tracker) AdvanceGame(gameID, move, nextFEN string) (CursorSnapshot, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	cursor, ok := t.games[gameID]
	if !ok {
		return CursorSnapshot{}, ErrGameNotTracked
	}

	t.ensureNodeLocked(cursor.currentHash, cursor.currentFEN)

	nextHash := position.HashFEN(nextFEN)
	nextNode := t.ensureNodeLocked(nextHash, nextFEN)
	currentNode := t.nodes[cursor.currentHash]
	currentNode.Children[normalizeMove(move)] = nextNode.PositionHash

	cursor.currentHash = nextHash
	cursor.currentFEN = nextFEN
	cursor.ply++

	return CursorSnapshot{
		GameID:              gameID,
		RootPositionHash:    cursor.rootHash,
		CurrentPositionHash: cursor.currentHash,
		Ply:                 cursor.ply,
	}, nil
}

func (t *Tracker) Snapshot(gameID string) (CursorSnapshot, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	cursor, ok := t.games[gameID]
	if !ok {
		return CursorSnapshot{}, false
	}

	return CursorSnapshot{
		GameID:              gameID,
		RootPositionHash:    cursor.rootHash,
		CurrentPositionHash: cursor.currentHash,
		Ply:                 cursor.ply,
	}, true
}

func (t *Tracker) ForgetGame(gameID string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	delete(t.games, gameID)
}

func (t *Tracker) ensureNodeLocked(hash, fen string) *PositionNode {
	if node, ok := t.nodes[hash]; ok {
		return node
	}

	node := &PositionNode{
		PositionHash: hash,
		FEN:          fen,
		Children:     make(map[string]string),
	}
	t.nodes[hash] = node
	return node
}

func normalizeMove(move string) string {
	return strings.ToLower(strings.TrimSpace(move))
}
