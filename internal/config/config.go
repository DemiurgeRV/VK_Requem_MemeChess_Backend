package config

import "os"

type Config struct {
	HTTPPort  string
	JWTSecret string
}

func Load() Config {
	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "8080"
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "super-secret-key"
	}

	return Config{
		HTTPPort:  port,
		JWTSecret: secret,
	}
}
