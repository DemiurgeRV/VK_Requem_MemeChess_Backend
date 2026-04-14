package service

import (
	"meme_chess/internal/analyzer/position"
	"sync"
)

type precomputeTask struct {
	state           *position.GameState
	remainingLevels int
}

type precomputeResult struct {
	hash            string
	remainingLevels int
	children        []precomputeTask
	err             error
}

// precomputeTree walks the position graph breadth-first so callers can
// progressively warm a bounded neighborhood around a root position.
func (s *AnalyzerService) precomputeTree(gs *position.GameState, depth int, treeDepth int) error {
	frontier := []precomputeTask{{
		state:           gs.Clone(),
		remainingLevels: treeDepth,
	}}

	seen := map[string]int{}
	computed := map[string]int{}

	for len(frontier) > 0 {
		nextFrontier, levelComputed, err := s.processPrecomputeLevel(frontier, depth, seen)
		if err != nil {
			return err
		}
		for hash, remainingLevels := range levelComputed {
			if current, ok := computed[hash]; !ok || remainingLevels > current {
				computed[hash] = remainingLevels
			}
		}
		frontier = nextFrontier
	}

	for hash, remainingLevels := range computed {
		s.cache.SetTreeDepth(hash, remainingLevels)
		s.cache.SetFrontierDepth(hash, remainingLevels)
	}

	return nil
}

func (s *AnalyzerService) processPrecomputeLevel(frontier []precomputeTask, depth int, seen map[string]int) ([]precomputeTask, map[string]int, error) {
	workers := s.precomputeWorkers
	if workers < 1 {
		workers = 1
	}

	jobs := make(chan precomputeTask)
	results := make(chan precomputeResult, len(frontier))
	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range jobs {
				results <- s.handlePrecomputeTask(task, depth)
			}
		}()
	}

	for _, task := range frontier {
		hash := task.state.Hash()
		if current, ok := seen[hash]; ok && current >= task.remainingLevels {
			continue
		}
		seen[hash] = task.remainingLevels
		jobs <- task
	}
	close(jobs)

	wg.Wait()
	close(results)

	nextFrontier := make([]precomputeTask, 0)
	levelComputed := make(map[string]int)
	var firstErr error

	for result := range results {
		if result.err != nil && firstErr == nil {
			firstErr = result.err
			continue
		}
		if current, ok := levelComputed[result.hash]; !ok || result.remainingLevels > current {
			levelComputed[result.hash] = result.remainingLevels
		}
		nextFrontier = append(nextFrontier, result.children...)
	}

	return nextFrontier, levelComputed, firstErr
}

func (s *AnalyzerService) handlePrecomputeTask(task precomputeTask, depth int) precomputeResult {
	hash := task.state.Hash()
	if pa, ok := s.cache.GetPosition(hash); ok && pa.Depth >= depth && pa.Ready && pa.TreeDepth >= task.remainingLevels {
		return precomputeResult{hash: hash, remainingLevels: task.remainingLevels}
	}

	if err := s.analyzeAndStorePosition(task.state, depth); err != nil {
		return precomputeResult{hash: hash, remainingLevels: task.remainingLevels, err: err}
	}

	result := precomputeResult{hash: hash, remainingLevels: task.remainingLevels}
	if task.remainingLevels <= 0 {
		return result
	}

	for _, mv := range s.gen.GenerateLegalMoves(task.state) {
		next := task.state.Clone()
		if err := next.ApplyMove(mv); err != nil {
			continue
		}
		result.children = append(result.children, precomputeTask{
			state:           next,
			remainingLevels: task.remainingLevels - 1,
		})
	}

	return result
}
