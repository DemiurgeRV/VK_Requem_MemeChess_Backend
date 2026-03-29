package config

import "os"

type Config struct {
	HTTPPort    string
	JWTSecret   string
	PostgresDSN string
}

func Load() Config {
	return Config{
		HTTPPort:    getEnv("HTTP_PORT", "8080"),
		JWTSecret:   getEnv("JWT_SECRET", "super-secret-dev-key"),
		PostgresDSN: getEnv("POSTGRES_DSN", "postgres://memechess:memechess@localhost:5432/meme_chess?sslmode=disable"),
	}
}

func getEnv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}
