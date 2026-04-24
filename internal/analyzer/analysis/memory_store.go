package analysis

import (
	"sync"

	"meme_chess/internal/analyzer/pattern"
)

type MemoryStore struct {
	mu        sync.RWMutex
	positions map[string]*PositionAnalysis
}

func NewCache() *MemoryStore {
	return &MemoryStore{
		positions: make(map[string]*PositionAnalysis),
	}
}

func (c *MemoryStore) GetPosition(hash string) (*PositionAnalysis, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	p, ok := c.positions[hash]
	if !ok {
		return nil, false
	}
	return p, true
}

func (c *MemoryStore) PutPosition(p *PositionAnalysis) {
	if p == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	current, ok := c.positions[p.PositionHash]
	if !ok {
		c.positions[p.PositionHash] = clonePositionAnalysis(p)
		return
	}

	if p.Depth > current.Depth {
		current.Depth = p.Depth
	}
	if p.TreeDepth > current.TreeDepth {
		current.TreeDepth = p.TreeDepth
	}
	if p.FrontierDepth > current.FrontierDepth {
		current.FrontierDepth = p.FrontierDepth
	}
	current.BestScoreCP = p.BestScoreCP
	current.Ready = p.Ready
}

func (c *MemoryStore) SetTreeDepth(hash string, treeDepth int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if p, ok := c.positions[hash]; ok && treeDepth > p.TreeDepth {
		p.TreeDepth = treeDepth
	}
}

func (c *MemoryStore) SetFrontierDepth(hash string, frontierDepth int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if p, ok := c.positions[hash]; ok && frontierDepth > p.FrontierDepth {
		p.FrontierDepth = frontierDepth
	}
}

func (c *MemoryStore) GetMove(hash, moveKey string) (*MoveAnalysis, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	p, ok := c.positions[hash]
	if !ok {
		return nil, false
	}
	move, ok := p.Moves[moveKey]
	if !ok {
		return nil, false
	}
	return move, true
}

func (c *MemoryStore) PutMove(hash string, depth int, bestScore int, moveKey string, move *MoveAnalysis) {
	if move == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	p, ok := c.positions[hash]
	if !ok {
		p = &PositionAnalysis{
			PositionHash: hash,
			Moves:        make(map[string]*MoveAnalysis),
		}
		c.positions[hash] = p
	}

	if depth > p.Depth {
		p.Depth = depth
	}
	p.BestScoreCP = bestScore
	p.Ready = true
	if p.Moves == nil {
		p.Moves = make(map[string]*MoveAnalysis)
	}

	current, ok := p.Moves[moveKey]
	if !ok {
		p.Moves[moveKey] = cloneMoveAnalysis(move)
		return
	}

	current.ScoreCP = move.ScoreCP
	current.DeltaCP = move.DeltaCP
	current.Quality = move.Quality
	current.Tags = append([]pattern.Tag(nil), move.Tags...)
	if move.Depth > current.Depth {
		current.Depth = move.Depth
	}
	current.NextPositionHash = move.NextPositionHash
	current.Ready = move.Ready
}

func (c *MemoryStore) Stats() (positions int, moves int) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	positions = len(c.positions)
	for _, p := range c.positions {
		moves += len(p.Moves)
	}
	return positions, moves
}

func (c *MemoryStore) Dump() map[string]any {
	c.mu.RLock()
	defer c.mu.RUnlock()

	out := make(map[string]any, len(c.positions))
	for hash, p := range c.positions {
		moves := make(map[string]any, len(p.Moves))
		for moveKey, move := range p.Moves {
			moves[moveKey] = map[string]any{
				"score_cp":           move.ScoreCP,
				"delta_cp":           move.DeltaCP,
				"quality":            move.Quality,
				"tags":               append([]pattern.Tag(nil), move.Tags...),
				"depth":              move.Depth,
				"next_position_hash": move.NextPositionHash,
				"ready":              move.Ready,
			}
		}

		out[hash] = map[string]any{
			"depth":          p.Depth,
			"tree_depth":     p.TreeDepth,
			"frontier_depth": p.FrontierDepth,
			"best_score":     p.BestScoreCP,
			"ready":          p.Ready,
			"moves":          moves,
		}
	}
	return out
}

func clonePositionAnalysis(p *PositionAnalysis) *PositionAnalysis {
	if p == nil {
		return nil
	}

	cp := &PositionAnalysis{
		PositionHash:  p.PositionHash,
		Depth:         p.Depth,
		TreeDepth:     p.TreeDepth,
		FrontierDepth: p.FrontierDepth,
		BestScoreCP:   p.BestScoreCP,
		Ready:         p.Ready,
		Moves:         make(map[string]*MoveAnalysis, len(p.Moves)),
	}
	for moveKey, move := range p.Moves {
		cp.Moves[moveKey] = cloneMoveAnalysis(move)
	}
	return cp
}

func cloneMoveAnalysis(move *MoveAnalysis) *MoveAnalysis {
	if move == nil {
		return nil
	}

	cp := *move
	cp.Tags = append(cp.Tags[:0:0], move.Tags...)
	return &cp
}
