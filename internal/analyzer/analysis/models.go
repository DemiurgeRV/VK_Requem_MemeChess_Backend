package analysis

import (
	"meme_chess/internal/analyzer/pattern"
	"meme_chess/internal/analyzer/position"
)

type MoveAnalysis struct {
	Move             position.Move `json:"-"`
	ScoreCP          int           `json:"score_cp"`
	DeltaCP          int           `json:"delta_cp"`
	Quality          string        `json:"quality"`
	Tags             []pattern.Tag `json:"tags"`
	Depth            int           `json:"depth"`
	NextPositionHash string        `json:"next_position_hash,omitempty"`
	Ready            bool          `json:"ready"`
}

type PositionAnalysis struct {
	PositionHash  string
	Depth         int
	TreeDepth     int
	FrontierDepth int
	BestScoreCP   int
	Moves         map[string]*MoveAnalysis
	Ready         bool
}

type Result struct {
	Move      string        `json:"move"`
	ScoreCP   int           `json:"score_cp"`
	DeltaCP   int           `json:"delta_cp"`
	Quality   string        `json:"quality"`
	Tags      []pattern.Tag `json:"tags"`
	Depth     int           `json:"depth"`
	FromCache bool          `json:"from_cache"`
}
