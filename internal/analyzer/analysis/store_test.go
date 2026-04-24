package analysis_test

import (
	"meme_chess/internal/analyzer/analysis"
	"meme_chess/internal/analyzer/pattern"
	"testing"
)

func TestMemoryStore_PersistsMovesWithinInstance(t *testing.T) {
	store := analysis.NewCache()

	store.PutMove("pos-1", 4, 27, "e2e4", &analysis.MoveAnalysis{
		ScoreCP:          27,
		DeltaCP:          0,
		Quality:          "best",
		Tags:             []pattern.Tag{pattern.TagCheck},
		Depth:            4,
		NextPositionHash: "pos-2",
		Ready:            true,
	})
	store.SetTreeDepth("pos-1", 2)
	store.SetFrontierDepth("pos-1", 3)

	move, ok := store.GetMove("pos-1", "e2e4")
	if !ok {
		t.Fatal("expected move to be loaded from store")
	}

	if move.ScoreCP != 27 || move.Depth != 4 {
		t.Fatalf("unexpected move payload: %+v", move)
	}
	if move.NextPositionHash != "pos-2" {
		t.Fatalf("unexpected next position hash: %s", move.NextPositionHash)
	}
	if len(move.Tags) != 1 || move.Tags[0] != pattern.TagCheck {
		t.Fatalf("unexpected tags: %+v", move.Tags)
	}

	position, ok := store.GetPosition("pos-1")
	if !ok {
		t.Fatal("expected position to be loaded from store")
	}
	if position.TreeDepth != 2 {
		t.Fatalf("unexpected tree depth: got %d want 2", position.TreeDepth)
	}
	if position.FrontierDepth != 3 {
		t.Fatalf("unexpected frontier depth: got %d want 3", position.FrontierDepth)
	}
}

func TestMemoryStore_DumpIncludesMoves(t *testing.T) {
	store := analysis.NewCache()
	store.PutMove("pos-1", 3, 12, "g1f3", &analysis.MoveAnalysis{
		ScoreCP: 12,
		DeltaCP: 5,
		Quality: "good",
		Depth:   3,
		Ready:   true,
	})

	dump := store.Dump()
	entry, ok := dump["pos-1"].(map[string]any)
	if !ok {
		t.Fatalf("expected position dump entry, got %#v", dump["pos-1"])
	}

	moves, ok := entry["moves"].(map[string]any)
	if !ok {
		t.Fatalf("expected moves map, got %#v", entry["moves"])
	}
	if _, ok := moves["g1f3"]; !ok {
		t.Fatalf("expected g1f3 to be present in dump, got %#v", moves)
	}
}
