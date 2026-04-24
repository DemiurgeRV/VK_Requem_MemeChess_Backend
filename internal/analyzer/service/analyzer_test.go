package service_test

import (
	"meme_chess/internal/analyzer/analysis"
	"meme_chess/internal/analyzer/pattern"
	"meme_chess/internal/analyzer/position"
	"meme_chess/internal/analyzer/rules"
	"meme_chess/internal/analyzer/service"
	"testing"
	"time"
)

func TestAnalyzer_AnalyzeMove_WithoutWarmup_ComputesAndReturnsResult(t *testing.T) {
	rs := rules.NewClassicalRuleSet()
	cache := analysis.NewCache()
	svc := service.NewAnalyzerService(rs, cache)

	gs := newGame(t,
		"e2e4",
		"e7e5",
	)

	mv := mustMove(t, "g1f3")

	res, err := svc.AnalyzeMove(gs, mv, 3)
	if err != nil {
		t.Fatalf("AnalyzeMove failed: %v", err)
	}

	if res == nil {
		t.Fatal("result is nil")
	}

	if res.Move != "g1f3" {
		t.Fatalf("unexpected move key: got %s want g1f3", res.Move)
	}

	if res.Depth != 3 {
		t.Fatalf("unexpected depth: got %d want 3", res.Depth)
	}
}

func TestAnalyzer_AnalyzeMove_IllegalMove_ReturnsError(t *testing.T) {
	rs := rules.NewClassicalRuleSet()
	cache := analysis.NewCache()
	svc := service.NewAnalyzerService(rs, cache)

	gs := newGame(t)

	// e2e5 из стартовой позиции нелегален
	mv := mustMove(t, "e2e5")

	_, err := svc.AnalyzeMove(gs, mv, 3)
	if err == nil {
		t.Fatal("expected illegal move error, got nil")
	}
}

func TestAnalyzer_WarmupPosition_FillsCache(t *testing.T) {
	rs := rules.NewClassicalRuleSet()
	cache := analysis.NewCache()
	svc := service.NewAnalyzerService(rs, cache)

	gs := newGame(t,
		"e2e4",
		"e7e5",
	)

	if err := svc.WarmupPosition(gs, 3); err != nil {
		t.Fatalf("WarmupPosition failed: %v", err)
	}

	positions, moves := svc.CacheStats()

	if positions == 0 {
		t.Fatalf("expected cached positions > 0, got %d", positions)
	}

	if moves == 0 {
		t.Fatalf("expected cached moves > 0, got %d", moves)
	}
}

func TestAnalyzer_AnalyzeMove_ReusesPreparedResult(t *testing.T) {
	rs := rules.NewClassicalRuleSet()
	cache := analysis.NewCache()
	svc := service.NewAnalyzerService(rs, cache)

	gs := newGame(t,
		"d2d4",
		"d7d5",
	)

	if err := svc.WarmupPosition(gs, 3); err != nil {
		t.Fatalf("warmup failed: %v", err)
	}

	mv := mustMove(t, "c1g5")

	res, err := svc.AnalyzeMove(gs, mv, 3)
	if err != nil {
		t.Fatalf("AnalyzeMove failed: %v", err)
	}

	if res == nil {
		t.Fatal("result is nil")
	}

	if !res.FromCache {
		t.Fatalf("expected prepared move to be returned from cache")
	}
}

func TestAnalyzer_BadMoveGetsSeverityTags(t *testing.T) {
	rs := rules.NewClassicalRuleSet()
	cache := analysis.NewCache()
	svc := service.NewAnalyzerService(rs, cache)

	gs := newGame(t,
		"e2e4",
		"e7e5",
		"d1h5",
		"b8c6",
	)

	if err := svc.WarmupPosition(gs, 3); err != nil {
		t.Fatalf("warmup failed: %v", err)
	}

	mv := mustMove(t, "h5e5")
	res, err := svc.AnalyzeMove(gs, mv, 3)
	if err != nil {
		t.Fatalf("analyze bad move failed: %v", err)
	}

	if len(res.Tags) == 0 {
		t.Fatalf("expected severity tags, got none")
	}
}

func TestAnalyzer_SearchQualityAccountsForOpponentReply(t *testing.T) {
	rs := rules.NewClassicalRuleSet()
	cache := analysis.NewCache()
	svc := service.NewAnalyzerService(rs, cache)

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

	mv := position.Move{From: position.MustSquare(3, 0), To: position.MustSquare(3, 6)}
	res, err := svc.AnalyzeMove(gs, mv, 2)
	if err != nil {
		t.Fatalf("AnalyzeMove failed: %v", err)
	}

	if res.Quality != "blunder" && res.Quality != "mistake" {
		t.Fatalf("expected poisoned capture to be punished, got quality %s score %d delta %d", res.Quality, res.ScoreCP, res.DeltaCP)
	}
	if !hasTag(res.Tags, pattern.TagMistake) && !hasTag(res.Tags, pattern.TagBlunder) {
		t.Fatalf("expected severity tag in %+v", res.Tags)
	}
}

func TestAnalyzer_WarmupTreePosition_PrecachesChildPositions(t *testing.T) {
	rs := rules.NewClassicalRuleSet()
	cache := analysis.NewCache()
	svc := service.NewAnalyzerService(rs, cache)
	svc.SetPrecomputeWorkers(2)

	gs := newGame(t)
	if err := svc.WarmupTreePosition(gs, 2, 1); err != nil {
		t.Fatalf("WarmupTreePosition failed: %v", err)
	}

	rootHash := gs.Hash()
	rootPos, ok := cache.GetPosition(rootHash)
	if !ok {
		t.Fatal("expected root position in cache")
	}
	if rootPos.TreeDepth < 1 {
		t.Fatalf("expected root tree depth >= 1, got %d", rootPos.TreeDepth)
	}

	rootMove, ok := cache.GetMove(rootHash, "e2e4")
	if !ok {
		t.Fatal("expected e2e4 move in root cache")
	}
	if rootMove.NextPositionHash == "" {
		t.Fatal("expected next position hash for e2e4")
	}

	childPos, ok := cache.GetPosition(rootMove.NextPositionHash)
	if !ok {
		t.Fatalf("expected child position %s to be precached", rootMove.NextPositionHash)
	}
	if childPos.Depth < 2 {
		t.Fatalf("expected child position depth >= 2, got %d", childPos.Depth)
	}
}

func TestAnalyzer_EnsureHotFrontier_PrecachesHumanMovesAndReplies(t *testing.T) {
	rs := rules.NewClassicalRuleSet()
	cache := analysis.NewCache()
	svc := service.NewAnalyzerService(rs, cache)
	svc.SetPrecomputeWorkers(2)

	gs := newGame(t)
	if err := svc.EnsureHotFrontier(gs, 2, 2); err != nil {
		t.Fatalf("EnsureHotFrontier failed: %v", err)
	}

	rootPos, ok := cache.GetPosition(gs.Hash())
	if !ok {
		t.Fatal("expected root position in cache")
	}
	if rootPos.FrontierDepth < 2 {
		t.Fatalf("expected frontier depth >= 2, got %d", rootPos.FrontierDepth)
	}

	e4, ok := cache.GetMove(gs.Hash(), "e2e4")
	if !ok || e4.NextPositionHash == "" {
		t.Fatal("expected precached e2e4 move with next position hash")
	}

	replyPos, ok := cache.GetPosition(e4.NextPositionHash)
	if !ok {
		t.Fatalf("expected human child position %s", e4.NextPositionHash)
	}
	if len(replyPos.Moves) == 0 {
		t.Fatal("expected engine replies to be precached from child position")
	}
}

func TestAnalyzer_EnsureHotFrontierAsync_QueuesWork(t *testing.T) {
	rs := rules.NewClassicalRuleSet()
	cache := analysis.NewCache()
	svc := service.NewAnalyzerService(rs, cache)
	svc.SetPrecomputeWorkers(2)

	gs := newGame(t)
	svc.EnsureHotFrontierAsync(gs, 2, 2)

	deadline := time.Now().Add(2 * time.Minute)
	for time.Now().Before(deadline) {
		if pos, ok := cache.GetPosition(gs.Hash()); ok && pos.FrontierDepth >= 2 {
			return
		}
		time.Sleep(25 * time.Millisecond)
	}

	pos, ok := cache.GetPosition(gs.Hash())
	if !ok {
		t.Fatal("expected position to be enqueued for frontier precompute")
	}
	t.Fatalf("expected frontier depth >= 2 after async queue, got %d", pos.FrontierDepth)
}

func hasTag(tags []pattern.Tag, want pattern.Tag) bool {
	for _, tag := range tags {
		if tag == want {
			return true
		}
	}
	return false
}
