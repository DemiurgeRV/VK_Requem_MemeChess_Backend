package game

import "meme_chess/internal/analyzer/analysis"

type MoveAnalyzer interface {
	StartGame(gameID string)
	RecordMove(gameID, move string)
	AnalyzeRecordedMove(gameID, move string, moveNumber int, depth int) (*analysis.Result, error)
	ForgetGame(gameID string)
}
