package service

import (
	"meme_chess/internal/analyzer/position"
	"sync"
)

type frontierRequest struct {
	state      *position.GameState
	depth      int
	horizonPly int
}

type frontierManager struct {
	service *AnalyzerService
	mu      sync.Mutex
	latest  *frontierRequest
	trigger chan struct{}
}

// frontierManager keeps only the newest requested frontier and computes it in
// the background, which is a good fit for interactive play where the current
// position may change faster than deep precompute can finish.
func newFrontierManager(service *AnalyzerService) *frontierManager {
	m := &frontierManager{
		service: service,
		trigger: make(chan struct{}, 1),
	}
	go m.loop()
	return m
}

func (m *frontierManager) Enqueue(gs *position.GameState, depth int, horizonPly int) {
	if gs == nil || depth < 1 || horizonPly < 0 {
		return
	}

	m.mu.Lock()
	m.latest = &frontierRequest{
		state:      gs,
		depth:      depth,
		horizonPly: horizonPly,
	}
	m.mu.Unlock()

	select {
	case m.trigger <- struct{}{}:
	default:
	}
}

func (m *frontierManager) loop() {
	for {
		<-m.trigger

		req := m.pop()
		if req == nil {
			continue
		}
		_ = m.service.EnsureHotFrontier(req.state, req.depth, req.horizonPly)
	}
}

func (m *frontierManager) pop() *frontierRequest {
	m.mu.Lock()
	defer m.mu.Unlock()

	req := m.latest
	m.latest = nil
	return req
}
