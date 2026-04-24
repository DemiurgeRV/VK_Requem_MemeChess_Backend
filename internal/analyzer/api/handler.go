package api

import (
	"encoding/json"
	"meme_chess/internal/analyzer/position"
	"meme_chess/internal/analyzer/service"
	"net/http"
	"strings"
)

const (
	defaultAnalysisDepth = 3
	defaultFrontierPly   = 2
)

type Handler struct {
	analyzer *service.AnalyzerService
}

func NewHandler(analyzer *service.AnalyzerService) *Handler {
	return &Handler{analyzer: analyzer}
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":      true,
		"service": "chess-analyzer",
	})
}

func (h *Handler) Metrics(w http.ResponseWriter, r *http.Request) {
	positions, moves := h.analyzer.CacheStats()

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":               true,
		"cached_positions": positions,
		"cached_moves":     moves,
	})
}

func (h *Handler) Warmup(w http.ResponseWriter, r *http.Request) {
	var req WarmupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	req.Depth = normalizeDepth(req.Depth)
	h.applyWorkers(req.Workers)

	gs, err := buildStateFromMoves(req.Moves)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if req.TreeDepth > 0 {
		err = h.analyzer.WarmupTreePosition(gs, req.Depth, req.TreeDepth)
	} else {
		err = h.analyzer.WarmupPosition(gs, req.Depth)
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, WarmupResponse{
		OK:      true,
		Message: "position prepared",
	})
}

func (h *Handler) Frontier(w http.ResponseWriter, r *http.Request) {
	var req FrontierRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	req.Depth = normalizeDepth(req.Depth)
	req.HorizonPly = normalizeFrontierPly(req.HorizonPly)
	h.applyWorkers(req.Workers)

	gs, err := buildStateFromMoves(req.Moves)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if req.Async {
		h.analyzer.EnsureHotFrontierAsync(gs, req.Depth, req.HorizonPly)
		writeJSON(w, http.StatusAccepted, map[string]any{
			"ok":      true,
			"message": "frontier queued",
		})
		return
	}

	if err := h.analyzer.EnsureHotFrontier(gs, req.Depth, req.HorizonPly); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":      true,
		"message": "frontier ready",
	})
}

func (h *Handler) AnalyzeMove(w http.ResponseWriter, r *http.Request) {
	var req AnalyzeMoveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	req.Depth = normalizeDepth(req.Depth)

	gs, err := buildStateFromMoves(req.Moves)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	mv, err := dtoToMove(gs, req.Move)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.analyzer.AnalyzeMove(gs, mv, req.Depth)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":     true,
		"result": result,
	})
}

func buildStateFromMoves(moves []MoveDTO) (*position.GameState, error) {
	gs := position.NewInitial()

	for _, dto := range moves {
		mv, err := dtoToMove(gs, dto)
		if err != nil {
			return nil, err
		}
		if err := gs.ApplyMove(mv); err != nil {
			return nil, err
		}
	}

	return gs, nil
}

func dtoToMove(gs *position.GameState, dto MoveDTO) (position.Move, error) {
	promotion := strings.TrimSpace(strings.ToLower(dto.Promotion))
	switch promotion {
	case "", "q", "r", "b", "n":
	default:
		return position.Move{}, httpError("invalid promotion")
	}

	return position.ParseUCIMove(gs, dto.From+dto.To+promotion)
}

type httpError string

func (e httpError) Error() string { return string(e) }

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, ErrorResponse{
		OK:    false,
		Error: msg,
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func (h *Handler) applyWorkers(workers int) {
	if workers > 0 {
		h.analyzer.SetPrecomputeWorkers(workers)
	}
}

func normalizeDepth(depth int) int {
	if depth > 0 {
		return depth
	}
	return defaultAnalysisDepth
}

func normalizeFrontierPly(ply int) int {
	if ply > 0 {
		return ply
	}
	return defaultFrontierPly
}

func (h *Handler) DebugCache(w http.ResponseWriter, r *http.Request) {
	data := h.analyzer.DumpCache()

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":    true,
		"cache": data,
	})
}
