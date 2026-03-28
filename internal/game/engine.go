package game

type MoveResult struct {
	FEN         string `json:"fen"`
	Move        string `json:"move"`
	IsCapture   bool   `json:"is_capture"`
	IsCheck     bool   `json:"is_check"`
	IsCheckmate bool   `json:"is_checkmate"`
}

type Engine interface {
	CurrentFEN() string
	ApplyMove(move string) (MoveResult, error)
}
