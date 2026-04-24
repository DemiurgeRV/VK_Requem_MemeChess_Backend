package rules

import "errors"

var (
	ErrNoPieceAtSource     = errors.New("no piece at source square")
	ErrWrongSideToMove     = errors.New("piece does not belong to side to move")
	ErrDestinationOccupied = errors.New("destination occupied by own piece")
	ErrIllegalGeometry     = errors.New("illegal move geometry")
	ErrKingLeftInCheck     = errors.New("move leaves king in check")
	ErrInvalidPromotion    = errors.New("invalid promotion")
	ErrIllegalCastle       = errors.New("illegal castle")
	ErrIllegalEnPassant    = errors.New("illegal en passant")
)
