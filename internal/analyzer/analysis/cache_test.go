package analysis_test

import (
	"meme_chess/internal/analyzer/analysis"
	"meme_chess/internal/analyzer/pattern"
	"os"
	"path/filepath"
	"testing"
)

func TestPersistentCache_PersistsMovesBetweenInstances(t *testing.T) {
	path := filepath.Join(t.TempDir(), "analysis.db")

	first, err := analysis.NewPersistentCache(path)
	if err != nil {
		t.Fatalf("create first cache: %v", err)
	}

	first.PutMove("pos-1", 4, 27, "e2e4", &analysis.MoveAnalysis{
		ScoreCP:          27,
		DeltaCP:          0,
		Quality:          "best",
		Tags:             []pattern.Tag{pattern.TagCheck},
		Depth:            4,
		NextPositionHash: "pos-2",
		Ready:            true,
	})
	first.SetTreeDepth("pos-1", 2)
	first.SetFrontierDepth("pos-1", 3)

	if err := first.Close(); err != nil {
		t.Fatalf("close first cache: %v", err)
	}

	second, err := analysis.NewPersistentCache(path)
	if err != nil {
		t.Fatalf("create second cache: %v", err)
	}
	defer second.Close()

	move, ok := second.GetMove("pos-1", "e2e4")
	if !ok {
		t.Fatal("expected move to be loaded from sqlite")
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

	position, ok := second.GetPosition("pos-1")
	if !ok {
		t.Fatal("expected position to be loaded from sqlite")
	}
	if position.TreeDepth != 2 {
		t.Fatalf("unexpected tree depth: got %d want 2", position.TreeDepth)
	}
	if position.FrontierDepth != 3 {
		t.Fatalf("unexpected frontier depth: got %d want 3", position.FrontierDepth)
	}
}

func TestPersistentCache_CreatesDatabaseFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "analysis.db")

	cache, err := analysis.NewPersistentCache(path)
	if err != nil {
		t.Fatalf("create cache: %v", err)
	}
	defer cache.Close()

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected sqlite file to exist: %v", err)
	}
}
