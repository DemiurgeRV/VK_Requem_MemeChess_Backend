package game

import (
	"strings"

	"github.com/notnil/chess"
)

type ChessEngine struct {
	game     *chess.Game
	notation chess.UCINotation
}

func NewChessEngine() *ChessEngine {
	return &ChessEngine{
		game:     chess.NewGame(),
		notation: chess.UCINotation{},
	}
}

func (e *ChessEngine) CurrentFEN() string {
	return e.game.Position().String()
}

func (e *ChessEngine) ApplyMove(uciMove string) (MoveResult, error) {
	uciMove = strings.TrimSpace(strings.ToLower(uciMove))
	if uciMove == "" {
		return MoveResult{}, ErrInvalidMove
	}

	beforePos := e.game.Position()
	beforeBoard := beforePos.Board()

	move, err := e.notation.Decode(beforePos, uciMove)
	if err != nil {
		return MoveResult{}, ErrInvalidMove
	}

	san := chess.AlgebraicNotation{}.Encode(beforePos, move)

	isCapture := beforeBoard.Piece(move.S2()) != chess.NoPiece
	isCheck := strings.HasSuffix(san, "+")
	isCheckmateByNotation := strings.HasSuffix(san, "#")

	if err := e.game.Move(move); err != nil {
		return MoveResult{}, ErrInvalidMove
	}

	fen := e.game.Position().String()

	outcome := e.game.Outcome()
	method := e.game.Method()

	isCheckmate := isCheckmateByNotation || (outcome != chess.NoOutcome && method == chess.Checkmate)

	return MoveResult{
		FEN:         fen,
		Move:        uciMove,
		IsCapture:   isCapture,
		IsCheck:     isCheck,
		IsCheckmate: isCheckmate,
	}, nil
}
