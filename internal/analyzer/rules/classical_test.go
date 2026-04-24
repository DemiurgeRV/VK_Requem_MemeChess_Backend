package rules

import (
	"meme_chess/internal/analyzer/position"
	"testing"
)

func TestPawnMoveLegal(t *testing.T) {
	gs := position.NewInitial()
	rs := NewClassicalRuleSet()

	mv, _ := position.NewMove("e2", "e4", position.NoPieceType)
	if err := rs.IsLegalMove(gs, mv); err != nil {
		t.Fatalf("expected legal move, got error: %v", err)
	}
}

func TestKnightMoveLegal(t *testing.T) {
	gs := position.NewInitial()
	rs := NewClassicalRuleSet()

	mv, _ := position.NewMove("g1", "f3", position.NoPieceType)
	if err := rs.IsLegalMove(gs, mv); err != nil {
		t.Fatalf("expected legal move, got error: %v", err)
	}
}

func TestBishopBlockedIllegal(t *testing.T) {
	gs := position.NewInitial()
	rs := NewClassicalRuleSet()

	mv, _ := position.NewMove("f1", "c4", position.NoPieceType)
	if err := rs.IsLegalMove(gs, mv); err == nil {
		t.Fatalf("expected illegal move because bishop path is blocked")
	}
}

func TestWrongSideToMove(t *testing.T) {
	gs := position.NewInitial()
	rs := NewClassicalRuleSet()

	mv, _ := position.NewMove("e7", "e5", position.NoPieceType)
	if err := rs.IsLegalMove(gs, mv); err == nil {
		t.Fatalf("expected wrong side to move error")
	}
}

func TestCannotCaptureOwnPiece(t *testing.T) {
	gs := position.NewInitial()
	rs := NewClassicalRuleSet()

	mv, _ := position.NewMove("e1", "d1", position.NoPieceType)
	if err := rs.IsLegalMove(gs, mv); err == nil {
		t.Fatalf("expected illegal move onto own piece")
	}
}
