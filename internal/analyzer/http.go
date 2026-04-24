package analyzer

import (
	"fmt"
	"io"
	"net/http"

	"meme_chess/internal/analyzer/analysis"
	"meme_chess/internal/analyzer/api"
	"meme_chess/internal/analyzer/rules"
	"meme_chess/internal/analyzer/service"
)

func NewHTTPHandler(cachePath string) (http.Handler, io.Closer, error) {
	cache, err := analysis.NewPersistentCache(cachePath)
	if err != nil {
		return nil, nil, fmt.Errorf("init analyzer cache: %w", err)
	}

	analyzerService := service.NewAnalyzerService(rules.NewClassicalRuleSet(), cache)
	handler := api.NewHandler(analyzerService)

	return api.NewRouter(handler), cache, nil
}
