package rules

import (
	"meme_chess/internal/analyzer/position"
	"testing"
)

func emptyBoard() *position.GameState {
	return &position.GameState{
		SideToMove:     position.White,
		EnPassant:      position.NoSquare,
		FullmoveNumber: 1,
	}
}

func TestWhiteKingSideCastleLegal(t *testing.T) {
	gs := emptyBoard()
	gs.SetPiece(position.MustSquare(4, 0), position.Piece{Type: position.King, Color: position.White})
	gs.SetPiece(position.MustSquare(7, 0), position.Piece{Type: position.Rook, Color: position.White})
	gs.SetPiece(position.MustSquare(4, 7), position.Piece{Type: position.King, Color: position.Black})
	gs.CastlingRights.WhiteKingSide = true

	rs := NewClassicalRuleSet()
	mv := position.Move{
		From: position.MustSquare(4, 0),
		To:   position.MustSquare(6, 0),
		Kind: position.MoveCastleKingSide,
	}

	if err := rs.IsLegalMove(gs, mv); err != nil {
		t.Fatalf("expected legal castle, got %v", err)
	}
}

func TestEnPassantLegal(t *testing.T) {
	gs := emptyBoard()
	gs.SetPiece(position.MustSquare(4, 4), position.Piece{Type: position.Pawn, Color: position.White}) // e5
	gs.SetPiece(position.MustSquare(3, 4), position.Piece{Type: position.Pawn, Color: position.Black}) // d5
	gs.SetPiece(position.MustSquare(4, 0), position.Piece{Type: position.King, Color: position.White})
	gs.SetPiece(position.MustSquare(4, 7), position.Piece{Type: position.King, Color: position.Black})
	gs.EnPassant = position.MustSquare(3, 5) // d6

	rs := NewClassicalRuleSet()
	mv := position.Move{
		From: position.MustSquare(4, 4),
		To:   position.MustSquare(3, 5),
		Kind: position.MoveEnPassant,
	}

	if err := rs.IsLegalMove(gs, mv); err != nil {
		t.Fatalf("expected legal en passant, got %v", err)
	}
}

func TestPromotionRequired(t *testing.T) {
	gs := emptyBoard()
	gs.SetPiece(position.MustSquare(0, 6), position.Piece{Type: position.Pawn, Color: position.White})
	gs.SetPiece(position.MustSquare(4, 0), position.Piece{Type: position.King, Color: position.White})
	gs.SetPiece(position.MustSquare(4, 7), position.Piece{Type: position.King, Color: position.Black})

	rs := NewClassicalRuleSet()
	mv := position.Move{
		From: position.MustSquare(0, 6),
		To:   position.MustSquare(0, 7),
	}

	if err := rs.IsLegalMove(gs, mv); err == nil {
		t.Fatalf("expected invalid promotion error")
	}
}
