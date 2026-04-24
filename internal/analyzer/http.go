package analyzer

import (
	"fmt"
	"net/http"

	"meme_chess/internal/analyzer/analysis"
	"meme_chess/internal/analyzer/api"
	"meme_chess/internal/analyzer/rules"
	"meme_chess/internal/analyzer/service"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewHTTPHandler(pool *pgxpool.Pool) (http.Handler, *service.GameTracker, error) {
	cache, err := analysis.NewPostgresStore(pool)
	if err != nil {
		return nil, nil, fmt.Errorf("init analyzer cache: %w", err)
	}

	analyzerService := service.NewAnalyzerService(rules.NewClassicalRuleSet(), cache)
	gameTracker := service.NewGameTracker(analyzerService)
	handler := api.NewHandler(analyzerService)

	return api.NewRouter(handler), gameTracker, nil
}
