package position

import "testing"

func TestParseUCIMove_RecognizesCastling(t *testing.T) {
	gs, err := BuildGameStateFromUCIMoves([]string{"e2e4", "e7e5", "g1f3", "b8c6", "f1e2", "g8f6"})
	if err != nil {
		t.Fatalf("build state: %v", err)
	}

	mv, err := ParseUCIMove(gs, "e1g1")
	if err != nil {
		t.Fatalf("parse move: %v", err)
	}
	if mv.Kind != MoveCastleKingSide {
		t.Fatalf("expected kingside castle, got %v", mv.Kind)
	}
}

func TestBuildGameStateFromUCIMoves_AppliesSpecialMoves(t *testing.T) {
	gs, err := BuildGameStateFromUCIMoves([]string{"e2e4", "a7a6", "e4e5", "d7d5", "e5d6"})
	if err != nil {
		t.Fatalf("build state: %v", err)
	}

	if piece := gs.PieceAt(MustSquare(3, 5)); piece.Type != Pawn || piece.Color != White {
		t.Fatalf("expected white pawn on d6 after en passant, got %+v", piece)
	}
	if captured := gs.PieceAt(MustSquare(3, 4)); !captured.IsZero() {
		t.Fatalf("expected captured pawn square to be empty, got %+v", captured)
	}
}
