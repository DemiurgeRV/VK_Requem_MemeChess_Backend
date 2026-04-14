package config

import (
	"os"
	"strings"
)

type Config struct {
	HTTPPort          string
	JWTSecret         string
	PostgresDSN       string
	FrontendJoinBase  string // e.g. http://localhost:5173; used to build /play/{game_id} links
	AnalysisCachePath string
}

func Load() Config {
	return Config{
		HTTPPort:          getEnv("HTTP_PORT", "8080"),
		JWTSecret:         getEnv("JWT_SECRET", "super-secret-dev-key"),
		PostgresDSN:       getEnv("POSTGRES_DSN", "postgres://memechess:memechess@localhost:5432/meme_chess?sslmode=disable"),
		FrontendJoinBase:  strings.TrimSuffix(getEnv("FRONTEND_JOIN_BASE", "http://localhost:5173"), "/"),
		AnalysisCachePath: getEnv("ANALYSIS_CACHE_PATH", "data/analysis.db"),
	}
}

func getEnv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}
