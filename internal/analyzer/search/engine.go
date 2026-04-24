package search

import (
	"meme_chess/internal/analyzer/movegen"
	"meme_chess/internal/analyzer/position"
	"meme_chess/internal/analyzer/rules"
)

const (
	negInf    = -10000000
	posInf    = 10000000
	MateScore = 1000000
)

type Engine struct {
	gen      *movegen.Generator
	rules    rules.RuleSet
	static   StaticEvaluator
	ordering MoveOrdering
	tt       TranspositionTable
}

func NewEngine(rs rules.RuleSet) *Engine {
	return &Engine{
		gen:      movegen.NewGenerator(rs),
		rules:    rs,
		static:   NewStaticEvaluator(),
		ordering: NewMoveOrdering(),
		tt:       NewTranspositionTable(),
	}
}

func (e *Engine) Analyze(gs *position.GameState, depth int) *Node {
	result := e.AnalyzePosition(gs, depth)

	children := make([]*Node, 0, len(result.RootMoves))
	for _, mv := range result.RootMoves {
		children = append(children, &Node{
			Hash:     gs.Hash(),
			Move:     mv.Move,
			Score:    mv.Score,
			Depth:    max(0, depth-1),
			Expanded: len(mv.PV) > 0,
			PV:       append([]position.Move(nil), mv.PV...),
		})
	}

	node := &Node{
		Hash:     result.Hash,
		Move:     result.BestMove,
		Score:    result.Score,
		Depth:    result.Depth,
		Expanded: true,
		PV:       append([]position.Move(nil), result.PV...),
		Children: children,
	}

	return node
}

// AnalyzePosition runs iterative deepening and returns one coherent root
// result: best move, principal variation and per-move scores for the entire
// position.
func (e *Engine) AnalyzePosition(gs *position.GameState, depth int) *Result {
	if depth < 1 {
		depth = 1
	}

	hash := gs.Hash()
	best := &Result{
		Hash:  hash,
		Depth: depth,
	}

	for currentDepth := 1; currentDepth <= depth; currentDepth++ {
		rootMoves := e.gen.GenerateLegalMoves(gs)
		if len(rootMoves) == 0 {
			score := e.terminalScore(gs, 0)
			best = &Result{
				Hash:      hash,
				Score:     score,
				Depth:     currentDepth,
				RootMoves: nil,
			}
			break
		}

		ttMove := position.NullMove()
		if entry, ok := e.tt.Get(hash); ok {
			ttMove = entry.BestMove
		}
		ordered := e.ordering.Order(gs, rootMoves, ttMove)

		alpha := negInf
		beta := posInf
		bestScore := negInf
		bestMove := position.NullMove()
		bestPV := []position.Move(nil)
		moveScores := make([]MoveScore, 0, len(ordered))
		nodes := 0

		for _, mv := range ordered {
			if err := gs.ApplyMove(mv); err != nil {
				continue
			}

			score, pv := e.negamax(gs, currentDepth-1, 1, -beta, -alpha, &nodes)
			score = -score

			if err := gs.UndoMove(); err != nil {
				panic(err)
			}

			line := append([]position.Move{mv}, pv...)
			moveScores = append(moveScores, MoveScore{
				Move:  mv,
				Score: score,
				PV:    line,
			})

			if score > bestScore {
				bestScore = score
				bestMove = mv
				bestPV = line
			}
			if score > alpha {
				alpha = score
			}
		}

		best = &Result{
			Hash:      hash,
			Score:     bestScore,
			BestMove:  bestMove,
			Depth:     currentDepth,
			PV:        bestPV,
			RootMoves: moveScores,
			Nodes:     nodes,
		}

		e.tt.Put(TTEntry{
			Hash:     hash,
			Depth:    currentDepth,
			Score:    bestScore,
			Bound:    BoundExact,
			BestMove: bestMove,
			PV:       append([]position.Move(nil), bestPV...),
		})
	}

	return best
}

func (e *Engine) negamax(gs *position.GameState, depth int, ply int, alpha, beta int, nodes *int) (int, []position.Move) {
	*nodes = *nodes + 1

	hash := gs.Hash()
	alphaOrig := alpha
	if entry, ok := e.tt.Get(hash); ok && entry.Depth >= depth {
		switch entry.Bound {
		case BoundExact:
			return entry.Score, append([]position.Move(nil), entry.PV...)
		case BoundLower:
			if entry.Score > alpha {
				alpha = entry.Score
			}
		case BoundUpper:
			if entry.Score < beta {
				beta = entry.Score
			}
		}
		if alpha >= beta {
			return entry.Score, append([]position.Move(nil), entry.PV...)
		}
	}

	if depth == 0 {
		return e.quiescence(gs, 0, alpha, beta, nodes), nil
	}

	moves := e.gen.GenerateLegalMoves(gs)
	if len(moves) == 0 {
		return e.terminalScore(gs, ply), nil
	}

	ttMove := position.NullMove()
	if entry, ok := e.tt.Get(hash); ok {
		ttMove = entry.BestMove
	}
	ordered := e.ordering.Order(gs, moves, ttMove)

	bestScore := negInf
	bestMove := position.NullMove()
	var bestPV []position.Move

	for _, mv := range ordered {
		if err := gs.ApplyMove(mv); err != nil {
			continue
		}

		score, childPV := e.negamax(gs, depth-1, ply+1, -beta, -alpha, nodes)
		score = -score

		if err := gs.UndoMove(); err != nil {
			panic(err)
		}

		if score > bestScore {
			bestScore = score
			bestMove = mv
			bestPV = append([]position.Move{mv}, childPV...)
		}
		if score > alpha {
			alpha = score
		}
		if alpha >= beta {
			break
		}
	}

	bound := BoundExact
	switch {
	case bestScore <= alphaOrig:
		bound = BoundUpper
	case bestScore >= beta:
		bound = BoundLower
	}

	e.tt.Put(TTEntry{
		Hash:     hash,
		Depth:    depth,
		Score:    bestScore,
		Bound:    bound,
		BestMove: bestMove,
		PV:       append([]position.Move(nil), bestPV...),
	})

	return bestScore, bestPV
}

func (e *Engine) terminalScore(gs *position.GameState, ply int) int {
	if e.rules.IsCheck(gs, gs.SideToMove) {
		return -MateScore + ply
	}
	return 0
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
