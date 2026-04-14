package service

import (
	"fmt"
	"meme_chess/internal/analyzer/analysis"
	"meme_chess/internal/analyzer/movegen"
	"meme_chess/internal/analyzer/pattern"
	"meme_chess/internal/analyzer/position"
	"meme_chess/internal/analyzer/rules"
	"meme_chess/internal/analyzer/search"
	"runtime"
)

type AnalyzerService struct {
	rules             rules.RuleSet
	gen               *movegen.Generator
	engine            *search.Engine
	detector          *pattern.Detector
	cache             *analysis.Cache
	precomputeWorkers int
	frontierManager   *frontierManager
}

func NewAnalyzerService(rs rules.RuleSet, cache *analysis.Cache) *AnalyzerService {
	workers := runtime.NumCPU()
	if workers < 2 {
		workers = 2
	}

	svc := &AnalyzerService{
		rules:             rs,
		gen:               movegen.NewGenerator(rs),
		engine:            search.NewEngine(rs),
		detector:          pattern.NewDetector(rs),
		cache:             cache,
		precomputeWorkers: workers,
	}
	svc.frontierManager = newFrontierManager(svc)
	return svc
}

func (s *AnalyzerService) WarmupPosition(gs *position.GameState, depth int) error {
	if gs == nil {
		return fmt.Errorf("nil game state")
	}
	if depth < 1 {
		return fmt.Errorf("depth must be >= 1")
	}

	hash := gs.Hash()
	if pa, ok := s.cache.GetPosition(hash); ok && pa.Depth >= depth && pa.Ready {
		return nil
	}

	return s.analyzeAndStorePosition(gs, depth)
}

func (s *AnalyzerService) WarmupTreePosition(gs *position.GameState, depth int, treeDepth int) error {
	if gs == nil {
		return fmt.Errorf("nil game state")
	}
	if depth < 1 {
		return fmt.Errorf("depth must be >= 1")
	}
	if treeDepth < 0 {
		return fmt.Errorf("tree depth must be >= 0")
	}
	if treeDepth == 0 {
		return s.WarmupPosition(gs, depth)
	}

	return s.precomputeTree(gs, depth, treeDepth)
}

func (s *AnalyzerService) EnsureHotFrontier(gs *position.GameState, depth int, horizonPly int) error {
	if gs == nil {
		return fmt.Errorf("nil game state")
	}
	if depth < 1 {
		return fmt.Errorf("depth must be >= 1")
	}
	if horizonPly < 0 {
		return fmt.Errorf("horizon ply must be >= 0")
	}

	hash := gs.Hash()
	if pa, ok := s.cache.GetPosition(hash); ok && pa.Depth >= depth && pa.FrontierDepth >= horizonPly && pa.Ready {
		return nil
	}

	if err := s.precomputeTree(gs, depth, horizonPly); err != nil {
		return err
	}
	s.cache.SetFrontierDepth(hash, horizonPly)
	return nil
}

func (s *AnalyzerService) EnsureHotFrontierAsync(gs *position.GameState, depth int, horizonPly int) {
	if gs == nil || s.frontierManager == nil {
		return
	}
	s.frontierManager.Enqueue(gs.Clone(), depth, horizonPly)
}

func (s *AnalyzerService) SetPrecomputeWorkers(workers int) {
	if workers < 1 {
		workers = 1
	}
	s.precomputeWorkers = workers
}

// analyzeAndStorePosition is the synchronous "single source of truth" path:
// one root search produces scores for every legal move, and those results are
// then written to SQLite as ready-to-serve move analyses.
func (s *AnalyzerService) analyzeAndStorePosition(gs *position.GameState, depth int) error {
	hash := gs.Hash()
	root := s.engine.AnalyzePosition(gs, depth)
	bestScore := root.Score

	for _, candidate := range root.RootMoves {
		delta := bestScore - candidate.Score
		tags, err := s.evaluateMove(gs, candidate, root, depth)
		if err != nil {
			continue
		}

		next := gs.Clone()
		if err := next.ApplyMove(candidate.Move); err != nil {
			continue
		}

		res := &analysis.MoveAnalysis{
			Move:             candidate.Move,
			ScoreCP:          candidate.Score,
			DeltaCP:          delta,
			Quality:          classifyDelta(delta),
			Tags:             dedupeTags(append(tags, qualityTags(delta)...)),
			Depth:            depth,
			NextPositionHash: next.Hash(),
			Ready:            true,
		}
		if isMissedOpportunity(bestScore, candidate.Score, delta) {
			res.Tags = dedupeTags(append(res.Tags, pattern.TagMissedOpportunity))
		}

		s.cache.PutMove(hash, depth, bestScore, moveKey(candidate.Move), res)
	}

	s.cache.SetTreeDepth(hash, 0)
	return nil
}

func (s *AnalyzerService) AnalyzeMove(gs *position.GameState, mv position.Move, depth int) (*analysis.Result, error) {
	if gs == nil {
		return nil, fmt.Errorf("nil game state")
	}
	if depth < 1 {
		return nil, fmt.Errorf("depth must be >= 1")
	}
	if err := s.rules.IsLegalMove(gs, mv); err != nil {
		return nil, err
	}

	hash := gs.Hash()
	key := moveKey(mv)
	if cached, ok := s.cache.GetMove(hash, key); ok && cached.Depth >= depth && cached.Ready {
		return &analysis.Result{
			Move:      key,
			ScoreCP:   cached.ScoreCP,
			DeltaCP:   cached.DeltaCP,
			Quality:   cached.Quality,
			Tags:      cached.Tags,
			Depth:     cached.Depth,
			FromCache: true,
		}, nil
	}

	if err := s.WarmupPosition(gs, depth); err != nil {
		return nil, err
	}

	cached, ok := s.cache.GetMove(hash, key)
	if !ok {
		return nil, fmt.Errorf("analysis was not prepared for move %s", key)
	}

	return &analysis.Result{
		Move:      key,
		ScoreCP:   cached.ScoreCP,
		DeltaCP:   cached.DeltaCP,
		Quality:   cached.Quality,
		Tags:      cached.Tags,
		Depth:     cached.Depth,
		FromCache: false,
	}, nil
}

func (s *AnalyzerService) CacheStats() (positions int, moves int) {
	return s.cache.Stats()
}

func (s *AnalyzerService) evaluateMove(gs *position.GameState, candidate search.MoveScore, root *search.Result, depth int) ([]pattern.Tag, error) {
	detectState := gs.Clone()
	tags, _ := s.detector.AnalyzeMove(detectState, candidate.Move, pattern.AnalysisContext{
		Move:      candidate,
		BestScore: root.Score,
		Delta:     root.Score - candidate.Score,
	})
	return dedupeTags(tags), nil
}

func moveKey(mv position.Move) string {
	if mv.Promotion != position.NoPieceType {
		return fmt.Sprintf("%s%s%s", mv.From, mv.To, promotionSuffix(mv.Promotion))
	}
	return fmt.Sprintf("%s%s", mv.From, mv.To)
}

func promotionSuffix(pt position.PieceType) string {
	switch pt {
	case position.Queen:
		return "q"
	case position.Rook:
		return "r"
	case position.Bishop:
		return "b"
	case position.Knight:
		return "n"
	default:
		return ""
	}
}

func classifyDelta(delta int) string {
	switch {
	case delta <= 30:
		return "best"
	case delta <= 90:
		return "good"
	case delta <= 170:
		return "inaccuracy"
	case delta <= 320:
		return "mistake"
	default:
		return "blunder"
	}
}

func qualityTags(delta int) []pattern.Tag {
	switch {
	case delta <= 90:
		return nil
	case delta <= 170:
		return []pattern.Tag{pattern.TagInaccuracy}
	case delta <= 320:
		return []pattern.Tag{pattern.TagMistake}
	default:
		return []pattern.Tag{pattern.TagBlunder}
	}
}

func isMissedOpportunity(bestScore, moveScore, delta int) bool {
	if delta < 140 {
		return false
	}
	return bestScore >= 180 || bestScore-moveScore >= 220
}

func dedupeTags(tags []pattern.Tag) []pattern.Tag {
	seen := make(map[pattern.Tag]struct{}, len(tags))
	out := make([]pattern.Tag, 0, len(tags))
	for _, t := range tags {
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		out = append(out, t)
	}
	return out
}

func (s *AnalyzerService) DumpCache() map[string]any {
	return s.cache.Dump()
}
