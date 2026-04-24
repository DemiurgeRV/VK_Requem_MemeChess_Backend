package position

import "testing"

func TestNewInitialFEN(t *testing.T) {
	gs := NewInitial()
	got := gs.FEN()
	want := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	if got != want {
		t.Fatalf("unexpected FEN\nwant: %s\ngot:  %s", want, got)
	}
}

func TestApplyMove(t *testing.T) {
	gs := NewInitial()
	mv, err := NewMove("e2", "e4", NoPieceType)
	if err != nil {
		t.Fatalf("NewMove error: %v", err)
	}

	if err := gs.ApplyMove(mv); err != nil {
		t.Fatalf("ApplyMove error: %v", err)
	}

	if piece := gs.PieceAt(MustSquare(4, 3)); piece.Type != Pawn || piece.Color != White {
		t.Fatalf("expected white pawn on e4, got %+v", piece)
	}
	if piece := gs.PieceAt(MustSquare(4, 1)); !piece.IsZero() {
		t.Fatalf("expected e2 to be empty, got %+v", piece)
	}
	if gs.SideToMove != Black {
		t.Fatalf("expected black to move, got %v", gs.SideToMove)
	}
}

func TestUndoMove(t *testing.T) {
	gs := NewInitial()
	before := gs.FEN()

	mv, _ := NewMove("e2", "e4", NoPieceType)
	if err := gs.ApplyMove(mv); err != nil {
		t.Fatalf("ApplyMove error: %v", err)
	}
	if err := gs.UndoMove(); err != nil {
		t.Fatalf("UndoMove error: %v", err)
	}

	after := gs.FEN()
	if before != after {
		t.Fatalf("state mismatch after undo\nwant: %s\ngot:  %s", before, after)
	}
}

func TestHashStableForEqualPositions(t *testing.T) {
	gs1 := NewInitial()
	gs2 := NewInitial()

	if gs1.Hash() != gs2.Hash() {
		t.Fatalf("expected equal hashes for equal positions")
	}
}

func TestMoveFields(t *testing.T) {
	mv, err := NewMove("a7", "a8", Queen)
	if err != nil {
		t.Fatalf("NewMove error: %v", err)
	}

	if mv.From.String() != "a7" {
		t.Fatalf("unexpected from: %s", mv.From)
	}
	if mv.To.String() != "a8" {
		t.Fatalf("unexpected to: %s", mv.To)
	}
	if mv.Promotion != Queen {
		t.Fatalf("unexpected promotion: %d", mv.Promotion)
	}
	if mv.Kind != MovePromotion {
		t.Fatalf("expected promotion kind, got %d", mv.Kind)
	}
}
