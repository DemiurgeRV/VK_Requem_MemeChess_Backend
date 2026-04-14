package movegen

import (
	"meme_chess/internal/analyzer/position"
	"meme_chess/internal/analyzer/rules"
	"testing"
)

func TestInitialPositionMoveCount(t *testing.T) {
	gs := position.NewInitial()
	rs := rules.NewClassicalRuleSet()
	gen := NewGenerator(rs)

	moves := gen.GenerateLegalMoves(gs)

	if len(moves) != 20 {
		t.Fatalf("expected 20 legal moves in initial position, got %d", len(moves))
	}
}

func TestMovesExistE2E4AndG1F3(t *testing.T) {
	gs := position.NewInitial()
	rs := rules.NewClassicalRuleSet()
	gen := NewGenerator(rs)

	moves := gen.GenerateLegalMoves(gs)

	if !containsMove(moves, "e2", "e4", position.NoPieceType) {
		t.Fatalf("expected move e2e4 to exist")
	}
	if !containsMove(moves, "g1", "f3", position.NoPieceType) {
		t.Fatalf("expected move g1f3 to exist")
	}
}

func containsMove(moves []position.Move, from, to string, promo position.PieceType) bool {
	fromSq, _ := position.ParseSquare(from)
	toSq, _ := position.ParseSquare(to)

	for _, mv := range moves {
		if mv.From == fromSq && mv.To == toSq && mv.Promotion == promo {
			return true
		}
	}

	return false
}
