package position

import "fmt"

type MoveKind uint8

const (
	MoveNormal MoveKind = iota
	MoveCastleKingSide
	MoveCastleQueenSide
	MoveEnPassant
	MovePromotion
)

type Move struct {
	From      Square
	To        Square
	Promotion PieceType
	Kind      MoveKind
}

func NullMove() Move {
	return Move{
		From: NoSquare,
		To:   NoSquare,
	}
}

func NewMove(from, to string, promotion PieceType) (Move, error) {
	fromSq, err := ParseSquare(from)
	if err != nil {
		return Move{}, fmt.Errorf("parse from square: %w", err)
	}

	toSq, err := ParseSquare(to)
	if err != nil {
		return Move{}, fmt.Errorf("parse to square: %w", err)
	}

	kind := MoveNormal
	if promotion != NoPieceType {
		kind = MovePromotion
	}

	return Move{
		From:      fromSq,
		To:        toSq,
		Promotion: promotion,
		Kind:      kind,
	}, nil
}

func (m Move) String() string {
	if m.Promotion != NoPieceType {
		return fmt.Sprintf("%s%s=%d", m.From, m.To, m.Promotion)
	}
	return fmt.Sprintf("%s%s", m.From, m.To)
}
