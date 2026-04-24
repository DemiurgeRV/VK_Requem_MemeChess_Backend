package search

import "meme_chess/internal/analyzer/position"

type Node struct {
	Hash     string
	Move     position.Move
	Score    int
	Depth    int
	Expanded bool
	PV       []position.Move

	Children []*Node
}

type MoveScore struct {
	Move  position.Move
	Score int
	PV    []position.Move
}

type Result struct {
	Hash      string
	Score     int
	BestMove  position.Move
	Depth     int
	PV        []position.Move
	RootMoves []MoveScore
	Nodes     int
}
