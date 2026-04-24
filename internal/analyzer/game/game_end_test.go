package game

import (
	"meme_chess/internal/analyzer/position"
	"meme_chess/internal/analyzer/rules"
	"testing"
)

func TestInitialPositionNotCheckmate(t *testing.T) {
	gs := position.NewInitial()
	rs := rules.NewClassicalRuleSet()

	if IsCheckmate(gs, rs) {
		t.Fatalf("initial position must not be checkmate")
	}
}

func TestInitialPositionNotStalemate(t *testing.T) {
	gs := position.NewInitial()
	rs := rules.NewClassicalRuleSet()

	if IsStalemate(gs, rs) {
		t.Fatalf("initial position must not be stalemate")
	}
}

func TestSimpleCheckmate(t *testing.T) {
	gs := &position.GameState{
		SideToMove:     position.Black,
		EnPassant:      position.NoSquare,
		FullmoveNumber: 1,
	}

	gs.SetPiece(position.MustSquare(7, 7), position.Piece{Type: position.King, Color: position.Black})  // h8
	gs.SetPiece(position.MustSquare(6, 6), position.Piece{Type: position.Queen, Color: position.White}) // g7
	gs.SetPiece(position.MustSquare(5, 5), position.Piece{Type: position.King, Color: position.White})  // f6

	rs := rules.NewClassicalRuleSet()

	if !IsCheckmate(gs, rs) {
		t.Fatalf("expected checkmate")
	}
}

func TestSimpleStalemate(t *testing.T) {
	gs := &position.GameState{
		SideToMove:     position.Black,
		EnPassant:      position.NoSquare,
		FullmoveNumber: 1,
	}

	gs.SetPiece(position.MustSquare(7, 7), position.Piece{Type: position.King, Color: position.Black})  // h8
	gs.SetPiece(position.MustSquare(5, 6), position.Piece{Type: position.Queen, Color: position.White}) // f7
	gs.SetPiece(position.MustSquare(6, 5), position.Piece{Type: position.King, Color: position.White})  // g6

	rs := rules.NewClassicalRuleSet()

	if !IsStalemate(gs, rs) {
		t.Fatalf("expected stalemate")
	}
}
