package pattern

import "meme_chess/internal/analyzer/search"

type AnalysisContext struct {
	Move      search.MoveScore
	BestScore int
	Delta     int
}
