package game

import (
	"meme_chess/internal/analyzer/movegen"
	"meme_chess/internal/analyzer/position"
	"meme_chess/internal/analyzer/rules"
)

func IsCheckmate(gs *position.GameState, rs rules.RuleSet) bool {
	gen := movegen.NewGenerator(rs)
	moves := gen.GenerateLegalMoves(gs)

	if len(moves) > 0 {
		return false
	}

	return rs.IsCheck(gs, gs.SideToMove)
}

func IsStalemate(gs *position.GameState, rs rules.RuleSet) bool {
	gen := movegen.NewGenerator(rs)
	moves := gen.GenerateLegalMoves(gs)

	if len(moves) > 0 {
		return false
	}

	return !rs.IsCheck(gs, gs.SideToMove)
}
