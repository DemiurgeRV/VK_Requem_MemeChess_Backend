package rules

import "meme_chess/internal/analyzer/position"

type RuleSet interface {
	IsLegalMove(gs *position.GameState, mv position.Move) error
	IsCheck(gs *position.GameState, color position.Color) bool
}
