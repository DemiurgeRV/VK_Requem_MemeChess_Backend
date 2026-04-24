package search

import (
	"meme_chess/internal/analyzer/position"
	"meme_chess/internal/analyzer/rules"
	"sort"
)

type MoveOrdering interface {
	Order(gs *position.GameState, moves []position.Move, ttMove position.Move) []position.Move
}

type DefaultMoveOrdering struct{}

func NewMoveOrdering() MoveOrdering {
	return DefaultMoveOrdering{}
}

func (DefaultMoveOrdering) Order(gs *position.GameState, moves []position.Move, ttMove position.Move) []position.Move {
	ordered := append([]position.Move(nil), moves...)

	sort.SliceStable(ordered, func(i, j int) bool {
		return movePriority(gs, ordered[i], ttMove) > movePriority(gs, ordered[j], ttMove)
	})

	return ordered
}

func movePriority(gs *position.GameState, mv position.Move, ttMove position.Move) int {
	score := 0

	if sameMove(mv, ttMove) {
		score += 50000
	}

	if captured := capturedPieceForMove(gs, mv); !captured.IsZero() {
		score += 10000 + pieceValue(captured.Type)*10 - pieceValue(gs.PieceAt(mv.From).Type)
	}

	switch mv.Kind {
	case position.MovePromotion:
		score += 9000 + pieceValue(mv.Promotion)
	case position.MoveEnPassant:
		score += 7000
	case position.MoveCastleKingSide, position.MoveCastleQueenSide:
		score += 500
	}

	if givesCheck(gs, mv) {
		score += 2500
	}

	// Mild centralization bonus helps move ordering without changing evaluation.
	to := mv.To
	score += 14 - abs(3-to.File()) - abs(3-to.Rank())

	return score
}

func givesCheck(gs *position.GameState, mv position.Move) bool {
	next := gs.Clone()
	if err := next.ApplyMove(mv); err != nil {
		return false
	}
	return rules.NewClassicalRuleSet().IsCheck(next, next.SideToMove)
}

func capturedPieceForMove(gs *position.GameState, mv position.Move) position.Piece {
	switch mv.Kind {
	case position.MoveEnPassant:
		mover := gs.PieceAt(mv.From)
		if mover.Color == position.White {
			return gs.PieceAt(position.MustSquare(mv.To.File(), mv.To.Rank()-1))
		}
		return gs.PieceAt(position.MustSquare(mv.To.File(), mv.To.Rank()+1))
	default:
		return gs.PieceAt(mv.To)
	}
}
