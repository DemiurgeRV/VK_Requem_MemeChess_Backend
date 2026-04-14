package search

import (
	"meme_chess/internal/analyzer/position"
	"meme_chess/internal/analyzer/rules"
	"testing"
)

func TestSearchInitial(t *testing.T) {
	gs := position.NewInitial()
	rs := rules.NewClassicalRuleSet()

	engine := NewEngine(rs)
	node := engine.Analyze(gs, 2)

	if node == nil {
		t.Fatalf("node is nil")
	}

	if node.Depth != 2 {
		t.Fatalf("expected depth 2, got %d", node.Depth)
	}
}

func TestSearchReturnsBestMove(t *testing.T) {
	gs := position.NewInitial()
	rs := rules.NewClassicalRuleSet()

	engine := NewEngine(rs)
	node := engine.Analyze(gs, 2)

	if node.Move.From == position.NoSquare && node.Move.To == position.NoSquare {
		t.Fatalf("expected best move to be set")
	}
}

func TestSearchDifferentDevelopingMovesGetDifferentScores(t *testing.T) {
	gs := position.NewInitial()
	rs := rules.NewClassicalRuleSet()
	engine := NewEngine(rs)

	for _, raw := range []struct {
		from string
		to   string
	}{
		{"e2", "e4"},
		{"e7", "e5"},
	} {
		mv, err := position.NewMove(raw.from, raw.to, position.NoPieceType)
		if err != nil {
			t.Fatalf("build move: %v", err)
		}
		if err := gs.ApplyMove(mv); err != nil {
			t.Fatalf("apply move: %v", err)
		}
	}

	nf3, _ := position.NewMove("g1", "f3", position.NoPieceType)
	na3, _ := position.NewMove("b1", "a3", position.NoPieceType)

	nf3State := gs.Clone()
	if err := nf3State.ApplyMove(nf3); err != nil {
		t.Fatalf("apply nf3: %v", err)
	}
	na3State := gs.Clone()
	if err := na3State.ApplyMove(na3); err != nil {
		t.Fatalf("apply na3: %v", err)
	}

	nf3Score := -engine.Analyze(nf3State, 2).Score
	na3Score := -engine.Analyze(na3State, 2).Score

	if nf3Score == na3Score {
		t.Fatalf("expected different scores for Nf3 and Na3, got %d and %d", nf3Score, na3Score)
	}
}

func TestSearchFindsMateInOne(t *testing.T) {
	gs := &position.GameState{
		SideToMove:     position.White,
		EnPassant:      position.NoSquare,
		FullmoveNumber: 1,
	}
	gs.SetPiece(position.MustSquare(5, 5), position.Piece{Type: position.King, Color: position.White})
	gs.SetPiece(position.MustSquare(6, 5), position.Piece{Type: position.Queen, Color: position.White})
	gs.SetPiece(position.MustSquare(7, 7), position.Piece{Type: position.King, Color: position.Black})

	engine := NewEngine(rules.NewClassicalRuleSet())
	result := engine.AnalyzePosition(gs, 2)

	want, _ := position.NewMove("g6", "g7", position.NoPieceType)
	if !sameMove(result.BestMove, want) {
		t.Fatalf("expected mate in one Qg7#, got %s", result.BestMove)
	}
	if result.Score < MateScore-10 {
		t.Fatalf("expected mating score, got %d", result.Score)
	}
}

func TestQuiescenceSeesImmediateRecapture(t *testing.T) {
	gs := &position.GameState{
		SideToMove:     position.White,
		EnPassant:      position.NoSquare,
		FullmoveNumber: 1,
	}
	gs.SetPiece(position.MustSquare(6, 0), position.Piece{Type: position.King, Color: position.White})
	gs.SetPiece(position.MustSquare(3, 0), position.Piece{Type: position.Queen, Color: position.White})
	gs.SetPiece(position.MustSquare(6, 7), position.Piece{Type: position.King, Color: position.Black})
	gs.SetPiece(position.MustSquare(3, 7), position.Piece{Type: position.Queen, Color: position.Black})
	gs.SetPiece(position.MustSquare(3, 6), position.Piece{Type: position.Rook, Color: position.Black})

	engine := NewEngine(rules.NewClassicalRuleSet())
	result := engine.AnalyzePosition(gs, 1)

	qxd7, _ := position.NewMove("d1", "d7", position.NoPieceType)
	score, ok := findMoveScore(result, qxd7)
	if !ok {
		t.Fatalf("expected Qxd7 to be legal and scored")
	}
	if score > -200 {
		t.Fatalf("expected quiescence to see recapture and keep Qxd7 poor, got %d", score)
	}
}

func findMoveScore(result *Result, move position.Move) (int, bool) {
	for _, candidate := range result.RootMoves {
		if sameMove(candidate.Move, move) {
			return candidate.Score, true
		}
	}
	return 0, false
}
