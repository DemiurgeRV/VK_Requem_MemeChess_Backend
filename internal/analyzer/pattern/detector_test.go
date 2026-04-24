package pattern

import (
	"meme_chess/internal/analyzer/position"
	"meme_chess/internal/analyzer/rules"
	"meme_chess/internal/analyzer/search"
	"testing"
)

func TestDetectSimpleCheck(t *testing.T) {
	gs := emptyBoard(position.White)
	gs.SetPiece(position.MustSquare(4, 0), position.Piece{Type: position.King, Color: position.White})
	gs.SetPiece(position.MustSquare(4, 7), position.Piece{Type: position.King, Color: position.Black})
	gs.SetPiece(position.MustSquare(3, 1), position.Piece{Type: position.Queen, Color: position.White})

	d := NewDetector(rules.NewClassicalRuleSet())
	mv := position.Move{From: position.MustSquare(3, 1), To: position.MustSquare(3, 6)}

	tags, _ := d.AnalyzeMove(gs, mv, ctxFor(mv))

	assertHasTag(t, tags, TagCheck)
}

func TestDetectFork(t *testing.T) {
	gs := emptyBoard(position.White)
	gs.SetPiece(position.MustSquare(4, 0), position.Piece{Type: position.King, Color: position.White})
	gs.SetPiece(position.MustSquare(4, 7), position.Piece{Type: position.King, Color: position.Black})
	gs.SetPiece(position.MustSquare(0, 7), position.Piece{Type: position.Rook, Color: position.Black})
	gs.SetPiece(position.MustSquare(3, 4), position.Piece{Type: position.Knight, Color: position.White})

	d := NewDetector(rules.NewClassicalRuleSet())
	mv := position.Move{From: position.MustSquare(3, 4), To: position.MustSquare(2, 6)}

	tags, _ := d.AnalyzeMove(gs, mv, ctxFor(mv))

	assertHasTag(t, tags, TagFork)
}

func TestDetectPinToKing(t *testing.T) {
	gs := emptyBoard(position.White)
	gs.SetPiece(position.MustSquare(4, 0), position.Piece{Type: position.King, Color: position.White})
	gs.SetPiece(position.MustSquare(4, 7), position.Piece{Type: position.King, Color: position.Black})
	gs.SetPiece(position.MustSquare(2, 5), position.Piece{Type: position.Knight, Color: position.Black})
	gs.SetPiece(position.MustSquare(2, 3), position.Piece{Type: position.Bishop, Color: position.White})

	d := NewDetector(rules.NewClassicalRuleSet())
	mv := position.Move{From: position.MustSquare(2, 3), To: position.MustSquare(1, 4)}

	tags, _ := d.AnalyzeMove(gs, mv, ctxFor(mv))

	assertHasTag(t, tags, TagPinToKing)
}

func TestDetectCheckmate(t *testing.T) {
	gs := emptyBoard(position.White)
	gs.SetPiece(position.MustSquare(6, 5), position.Piece{Type: position.King, Color: position.White})
	gs.SetPiece(position.MustSquare(6, 6), position.Piece{Type: position.Queen, Color: position.White})
	gs.SetPiece(position.MustSquare(7, 7), position.Piece{Type: position.King, Color: position.Black})

	d := NewDetector(rules.NewClassicalRuleSet())
	mv := position.Move{From: position.MustSquare(6, 6), To: position.MustSquare(7, 6)}

	tags, _ := d.AnalyzeMove(gs, mv, ctxFor(mv))

	assertHasTag(t, tags, TagCheckmate)
	assertHasTag(t, tags, TagForcedMate)
}

func TestDoesNotMarkSingleE4AsOpening(t *testing.T) {
	gs := position.NewInitial()
	d := NewDetector(rules.NewClassicalRuleSet())

	mv, _ := position.NewMove("e2", "e4", position.NoPieceType)
	tags, _ := d.AnalyzeMove(gs, mv, ctxFor(mv))

	assertNoTag(t, tags, TagOpening)
}

func TestMarksRecognizedOpeningSequence(t *testing.T) {
	gs := position.NewInitial()
	applyMoves(t, gs, "e2e4", "e7e5", "g1f3", "b8c6", "f1b5")

	tags := openingTags(gs)

	assertHasTag(t, tags, TagOpening)
	assertHasTag(t, tags, TagOpeningRuyLopez)
}

func emptyBoard(side position.Color) *position.GameState {
	return &position.GameState{
		SideToMove:     side,
		EnPassant:      position.NoSquare,
		FullmoveNumber: 1,
	}
}

func assertHasTag(t *testing.T, tags []Tag, want Tag) {
	t.Helper()
	for _, tag := range tags {
		if tag == want {
			return
		}
	}
	t.Fatalf("expected tag %s, got %+v", want, tags)
}

func assertNoTag(t *testing.T, tags []Tag, unwanted Tag) {
	t.Helper()
	for _, tag := range tags {
		if tag == unwanted {
			t.Fatalf("did not expect tag %s in %+v", unwanted, tags)
		}
	}
}

func applyMoves(t *testing.T, gs *position.GameState, moves ...string) {
	t.Helper()
	for _, raw := range moves {
		mv, err := position.NewMove(raw[0:2], raw[2:4], position.NoPieceType)
		if err != nil {
			t.Fatalf("build move %s: %v", raw, err)
		}
		if err := gs.ApplyMove(mv); err != nil {
			t.Fatalf("apply move %s: %v", raw, err)
		}
	}
}

func ctxFor(mv position.Move) AnalysisContext {
	return AnalysisContext{
		Move:      search.MoveScore{Move: mv, Score: 0},
		BestScore: 0,
		Delta:     0,
	}
}
