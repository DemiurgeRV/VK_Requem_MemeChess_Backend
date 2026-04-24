package api

type MoveDTO struct {
	From      string `json:"from"`
	To        string `json:"to"`
	Promotion string `json:"promotion,omitempty"`
}

type WarmupRequest struct {
	Moves     []MoveDTO `json:"moves"`
	Depth     int       `json:"depth"`
	TreeDepth int       `json:"tree_depth,omitempty"`
	Workers   int       `json:"workers,omitempty"`
}

type FrontierRequest struct {
	Moves      []MoveDTO `json:"moves"`
	Depth      int       `json:"depth"`
	HorizonPly int       `json:"horizon_ply,omitempty"`
	Workers    int       `json:"workers,omitempty"`
	Async      bool      `json:"async,omitempty"`
}

type AnalyzeMoveRequest struct {
	Moves []MoveDTO `json:"moves"`
	Move  MoveDTO   `json:"move"`
	Depth int       `json:"depth"`
}

type ErrorResponse struct {
	OK    bool   `json:"ok"`
	Error string `json:"error"`
}

type WarmupResponse struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}
