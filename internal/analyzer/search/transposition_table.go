package search

import (
	"meme_chess/internal/analyzer/position"
	"sync"
)

type Bound uint8

const (
	BoundExact Bound = iota
	BoundLower
	BoundUpper
)

type TTEntry struct {
	Hash     string
	Depth    int
	Score    int
	Bound    Bound
	BestMove position.Move
	PV       []position.Move
}

type TranspositionTable interface {
	Get(hash string) (TTEntry, bool)
	Put(entry TTEntry)
}

type MemoryTranspositionTable struct {
	mu      sync.RWMutex
	entries map[string]TTEntry
}

func NewTranspositionTable() *MemoryTranspositionTable {
	return &MemoryTranspositionTable{
		entries: make(map[string]TTEntry),
	}
}

func (t *MemoryTranspositionTable) Get(hash string) (TTEntry, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	entry, ok := t.entries[hash]
	return entry, ok
}

func (t *MemoryTranspositionTable) Put(entry TTEntry) {
	t.mu.Lock()
	defer t.mu.Unlock()

	current, ok := t.entries[entry.Hash]
	if ok && current.Depth > entry.Depth {
		return
	}

	t.entries[entry.Hash] = entry
}
