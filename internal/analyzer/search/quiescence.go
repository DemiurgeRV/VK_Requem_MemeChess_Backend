package search

import (
	"meme_chess/internal/analyzer/position"
	"meme_chess/internal/analyzer/rules"
)

const maxQuiescenceDepth = 8

func (e *Engine) quiescence(gs *position.GameState, ply int, alpha, beta int, nodes *int) int {
	*nodes = *nodes + 1
	standPat := e.static.Evaluate(gs)
	if standPat >= beta {
		return beta
	}
	if standPat > alpha {
		alpha = standPat
	}

	if ply >= maxQuiescenceDepth {
		return standPat
	}

	moves := e.noisyMoves(gs)
	if len(moves) == 0 {
		return standPat
	}

	ordered := e.ordering.Order(gs, moves, position.NullMove())
	for _, mv := range ordered {
		if err := gs.ApplyMove(mv); err != nil {
			continue
		}

		score := -e.quiescence(gs, ply+1, -beta, -alpha, nodes)

		if err := gs.UndoMove(); err != nil {
			panic(err)
		}

		if score >= beta {
			return beta
		}
		if score > alpha {
			alpha = score
		}
	}

	return alpha
}

func (e *Engine) noisyMoves(gs *position.GameState) []position.Move {
	all := e.gen.GenerateLegalMoves(gs)
	out := make([]position.Move, 0, len(all))

	for _, mv := range all {
		if e.isNoisyMove(gs, mv) {
			out = append(out, mv)
		}
	}

	return out
}

func (e *Engine) isNoisyMove(gs *position.GameState, mv position.Move) bool {
	if mv.Kind == position.MovePromotion || mv.Kind == position.MoveEnPassant {
		return true
	}
	if !capturedPieceForMove(gs, mv).IsZero() {
		return true
	}

	next := gs.Clone()
	if err := next.ApplyMove(mv); err != nil {
		return false
	}

	// Keep forcing continuations alive inside q-search.
	if e.rules.IsCheck(next, next.SideToMove) {
		return true
	}

	// Tactical threat: move attacks an undefended enemy piece.
	mover := next.SideToMove.Opponent()
	movedPiece := next.PieceAt(mv.To)
	for i := 0; i < 64; i++ {
		targetSq := position.Square(i)
		target := next.PieceAt(targetSq)
		if target.IsZero() || target.Color == mover {
			continue
		}
		if rules.AttacksSquare(next, mv.To, targetSq, movedPiece) && isWeakTarget(next, targetSq, target.Color) {
			return true
		}
	}

	return false
}
