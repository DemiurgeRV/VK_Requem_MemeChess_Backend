package movegen

import (
	"meme_chess/internal/analyzer/position"
	"meme_chess/internal/analyzer/rules"
	"testing"
)

func TestPerftInitial(t *testing.T) {
	gs := position.NewInitial()
	rs := rules.NewClassicalRuleSet()
	gen := NewGenerator(rs)

	tests := []struct {
		depth int
		nodes int
	}{
		{1, 20},
		{2, 400},
		{3, 8902},
	}

	for _, tt := range tests {
		got := Perft(gs, gen, tt.depth)
		if got != tt.nodes {
			t.Fatalf("depth %d: expected %d nodes, got %d", tt.depth, tt.nodes, got)
		}
	}
}
