package search

import (
	"meme_chess/internal/analyzer/moveeval"
	"meme_chess/internal/analyzer/position"
)

type StaticEvaluator interface {
	Evaluate(gs *position.GameState) int
}

type defaultStaticEvaluator struct{}

func NewStaticEvaluator() StaticEvaluator {
	return defaultStaticEvaluator{}
}

func (defaultStaticEvaluator) Evaluate(gs *position.GameState) int {
	return moveeval.Evaluate(gs)
}
