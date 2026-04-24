package analysis

type Store interface {
	GetPosition(hash string) (*PositionAnalysis, bool)
	PutPosition(p *PositionAnalysis)
	SetTreeDepth(hash string, treeDepth int)
	SetFrontierDepth(hash string, frontierDepth int)
	GetMove(hash, moveKey string) (*MoveAnalysis, bool)
	PutMove(hash string, depth int, bestScore int, moveKey string, move *MoveAnalysis)
	Stats() (positions int, moves int)
	Dump() map[string]any
}
