package position

import (
	"fmt"
	"strings"
)

func ParseUCIMove(gs *GameState, raw string) (Move, error) {
	raw = strings.TrimSpace(strings.ToLower(raw))
	if len(raw) != 4 && len(raw) != 5 {
		return Move{}, fmt.Errorf("invalid uci move: %q", raw)
	}

	from, err := ParseSquare(raw[:2])
	if err != nil {
		return Move{}, fmt.Errorf("parse from square: %w", err)
	}
	to, err := ParseSquare(raw[2:4])
	if err != nil {
		return Move{}, fmt.Errorf("parse to square: %w", err)
	}

	promotion := NoPieceType
	if len(raw) == 5 {
		switch raw[4] {
		case 'q':
			promotion = Queen
		case 'r':
			promotion = Rook
		case 'b':
			promotion = Bishop
		case 'n':
			promotion = Knight
		default:
			return Move{}, fmt.Errorf("invalid promotion: %q", string(raw[4]))
		}
	}

	move := Move{
		From:      from,
		To:        to,
		Promotion: promotion,
		Kind:      MoveNormal,
	}
	if promotion != NoPieceType {
		move.Kind = MovePromotion
	}

	if gs == nil {
		return move, nil
	}

	piece := gs.PieceAt(from)
	if piece.IsZero() {
		return Move{}, fmt.Errorf("no piece at source square %s", from)
	}

	if piece.Type == King && from.Rank() == to.Rank() && absInt(to.File()-from.File()) == 2 {
		if to.File() > from.File() {
			move.Kind = MoveCastleKingSide
		} else {
			move.Kind = MoveCastleQueenSide
		}
		move.Promotion = NoPieceType
	}

	if piece.Type == Pawn && move.Kind != MovePromotion && from.File() != to.File() && gs.PieceAt(to).IsZero() && gs.EnPassant == to {
		move.Kind = MoveEnPassant
	}

	return move, nil
}

func BuildGameStateFromUCIMoves(moves []string) (*GameState, error) {
	gs := NewInitial()
	for _, raw := range moves {
		mv, err := ParseUCIMove(gs, raw)
		if err != nil {
			return nil, err
		}
		if err := gs.ApplyMove(mv); err != nil {
			return nil, err
		}
	}
	return gs, nil
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
