package service_test

import (
	"meme_chess/internal/analyzer/position"
	"testing"
)

func newGame(t *testing.T, moves ...string) *position.GameState {
	t.Helper()

	gs := position.NewInitial()

	for _, raw := range moves {
		mv := mustMove(t, raw)
		if err := gs.ApplyMove(mv); err != nil {
			t.Fatalf("apply move %s failed: %v", raw, err)
		}
	}

	return gs
}

func mustMove(t *testing.T, uci string) position.Move {
	t.Helper()

	if len(uci) != 4 && len(uci) != 5 {
		t.Fatalf("invalid move string: %s", uci)
	}

	promo := position.NoPieceType
	if len(uci) == 5 {
		switch uci[4] {
		case 'q':
			promo = position.Queen
		case 'r':
			promo = position.Rook
		case 'b':
			promo = position.Bishop
		case 'n':
			promo = position.Knight
		default:
			t.Fatalf("invalid promotion suffix: %s", uci)
		}
	}

	mv, err := position.NewMove(uci[0:2], uci[2:4], promo)
	if err != nil {
		t.Fatalf("build move %s failed: %v", uci, err)
	}

	return mv
}
