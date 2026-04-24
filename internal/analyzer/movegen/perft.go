package movegen

import "meme_chess/internal/analyzer/position"

func Perft(gs *position.GameState, gen *Generator, depth int) int {
	if depth == 0 {
		return 1
	}

	nodes := 0
	moves := gen.GenerateLegalMoves(gs)

	for _, mv := range moves {
		if err := gs.ApplyMove(mv); err != nil {
			panic(err)
		}
		nodes += Perft(gs, gen, depth-1)
		if err := gs.UndoMove(); err != nil {
			panic(err)
		}
	}

	return nodes
}
