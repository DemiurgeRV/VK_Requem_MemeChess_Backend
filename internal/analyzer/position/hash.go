package position

import (
	"crypto/sha256"
	"encoding/hex"
)

func (g *GameState) Hash() string {
	return HashFEN(g.FEN())
}

func HashFEN(fen string) string {
	sum := sha256.Sum256([]byte(fen))
	return hex.EncodeToString(sum[:])
}
