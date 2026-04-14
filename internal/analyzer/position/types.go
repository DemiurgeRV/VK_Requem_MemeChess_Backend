package position

import "fmt"

type Color uint8

const (
	White Color = iota
	Black
)

func (c Color) Opponent() Color {
	if c == White {
		return Black
	}
	return White
}

func (c Color) String() string {
	if c == White {
		return "w"
	}
	return "b"
}

type PieceType uint8

const (
	NoPieceType PieceType = iota
	Pawn
	Knight
	Bishop
	Rook
	Queen
	King
)

type Piece struct {
	Type  PieceType
	Color Color
}

func (p Piece) IsZero() bool {
	return p.Type == NoPieceType
}

func (p Piece) FENChar() byte {
	var ch byte
	switch p.Type {
	case Pawn:
		ch = 'p'
	case Knight:
		ch = 'n'
	case Bishop:
		ch = 'b'
	case Rook:
		ch = 'r'
	case Queen:
		ch = 'q'
	case King:
		ch = 'k'
	default:
		return 0
	}

	if p.Color == White {
		return ch - 32
	}
	return ch
}

type Square uint8

const NoSquare Square = 255

func NewSquare(file, rank int) (Square, error) {
	if file < 0 || file > 7 || rank < 0 || rank > 7 {
		return NoSquare, fmt.Errorf("square out of range: file=%d rank=%d", file, rank)
	}
	return Square(rank*8 + file), nil
}

func MustSquare(file, rank int) Square {
	sq, err := NewSquare(file, rank)
	if err != nil {
		panic(err)
	}
	return sq
}

func ParseSquare(s string) (Square, error) {
	if len(s) != 2 {
		return NoSquare, fmt.Errorf("invalid square: %q", s)
	}

	file := int(s[0] - 'a')
	rank := int(s[1] - '1')
	return NewSquare(file, rank)
}

func (s Square) File() int {
	return int(s % 8)
}

func (s Square) Rank() int {
	return int(s / 8)
}

func (s Square) String() string {
	if s == NoSquare {
		return "-"
	}
	return fmt.Sprintf("%c%d", 'a'+s.File(), s.Rank()+1)
}
