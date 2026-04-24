package service_test

import (
	"meme_chess/internal/analyzer/analysis"
	"meme_chess/internal/analyzer/rules"
	"meme_chess/internal/analyzer/service"
	"testing"
)

func TestAnalyzer_WarmupThenAnalyzeThenCacheHit(t *testing.T) {
	rs := rules.NewClassicalRuleSet()
	cache := analysis.NewCache()
	svc := service.NewAnalyzerService(rs, cache)

	gs := newGame(t,
		"e2e4",
		"e7e5",
	)

	// 1) заранее прогреваем позицию
	if err := svc.WarmupPosition(gs, 3); err != nil {
		t.Fatalf("warmup failed: %v", err)
	}

	// 2) первый запрос анализа
	firstMove := mustMove(t, "g1f3")
	first, err := svc.AnalyzeMove(gs, firstMove, 3)
	if err != nil {
		t.Fatalf("first analyze failed: %v", err)
	}

	if first == nil {
		t.Fatal("first result is nil")
	}

	// после warmup ожидаем cache hit уже на первом запросе
	if !first.FromCache {
		t.Fatalf("expected first response to come from cache after warmup")
	}

	// 3) второй запрос того же хода
	second, err := svc.AnalyzeMove(gs, firstMove, 3)
	if err != nil {
		t.Fatalf("second analyze failed: %v", err)
	}

	if second == nil {
		t.Fatal("second result is nil")
	}

	if !second.FromCache {
		t.Fatalf("expected second response to come from cache")
	}

	// 4) проверяем, что результат стабилен
	if first.ScoreCP != second.ScoreCP {
		t.Fatalf("score mismatch: first=%d second=%d", first.ScoreCP, second.ScoreCP)
	}

	if first.DeltaCP != second.DeltaCP {
		t.Fatalf("delta mismatch: first=%d second=%d", first.DeltaCP, second.DeltaCP)
	}

	if first.Quality != second.Quality {
		t.Fatalf("quality mismatch: first=%s second=%s", first.Quality, second.Quality)
	}

	if first.Depth != second.Depth {
		t.Fatalf("depth mismatch: first=%d second=%d", first.Depth, second.Depth)
	}
}
