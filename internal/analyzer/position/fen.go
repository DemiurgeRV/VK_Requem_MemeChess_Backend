package position

import (
	"strconv"
	"strings"
)

func (g *GameState) FEN() string {
	var board strings.Builder
	for rank := 7; rank >= 0; rank-- {
		empty := 0
		for file := 0; file < 8; file++ {
			sq := rank*8 + file
			piece := g.Board[sq]
			if piece.IsZero() {
				empty++
				continue
			}
			if empty > 0 {
				board.WriteString(strconv.Itoa(empty))
				empty = 0
			}
			board.WriteByte(piece.FENChar())
		}
		if empty > 0 {
			board.WriteString(strconv.Itoa(empty))
		}
		if rank > 0 {
			board.WriteByte('/')
		}
	}

	ep := "-"
	if g.EnPassant != NoSquare {
		ep = g.EnPassant.String()
	}

	return strings.Join([]string{
		board.String(),
		g.SideToMove.String(),
		g.CastlingRights.String(),
		ep,
		strconv.Itoa(g.HalfmoveClock),
		strconv.Itoa(g.FullmoveNumber),
	}, " ")
}
