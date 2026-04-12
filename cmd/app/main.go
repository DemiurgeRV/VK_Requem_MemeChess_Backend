package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"meme_chess/internal/auth"
	"meme_chess/internal/config"
	"meme_chess/internal/db"
	"meme_chess/internal/game"
	"meme_chess/internal/user"
	"meme_chess/internal/ws"
)

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	cfg := config.Load()

	pg, err := db.NewPostgres(cfg.PostgresDSN)
	if err != nil {
		log.Fatalf("failed to connect postgres: %v", err)
	}

	jwtManager := auth.NewJWTManager(cfg.JWTSecret)
	userRepo := user.NewRepository(pg.Pool)
	authService := auth.NewService(userRepo, jwtManager)
	authHandlers := &auth.Handlers{Service: authService}

	hub := ws.NewHub()
	gameRepo := game.NewRepository(pg.Pool)
	gameService := game.NewService(gameRepo)
	wsHandler := ws.NewHandler(hub, gameService, jwtManager)

	go hub.Run()

	// DEBUG: создаём одну тестовую игру
	_, err = gameService.CreateGame(
		context.Background(),
		"11111111-1111-1111-1111-111111111111",
		"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
		"bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
		game.NewChessEngine(),
	)
	if err != nil {
		log.Printf("failed to create debug game: %v", err)
	}

	http.HandleFunc("/auth/register", authHandlers.Register)
	http.HandleFunc("/auth/login", authHandlers.Login)
	http.HandleFunc("/auth/me", authHandlers.Me)

	http.HandleFunc("/ws", wsHandler.ServeWS)

	http.HandleFunc("/debug/token", func(w http.ResponseWriter, r *http.Request) {
		userID := r.URL.Query().Get("user_id")
		if userID == "" {
			http.Error(w, "missing user_id", http.StatusBadRequest)
			return
		}

		token, err := jwtManager.Generate(userID)
		if err != nil {
			http.Error(w, "failed to generate token", http.StatusInternalServerError)
			return
		}

		_ = json.NewEncoder(w).Encode(map[string]string{
			"token": token,
		})
	})

	http.HandleFunc("/debug/game", func(w http.ResponseWriter, r *http.Request) {
		session, ok := gameService.GetSession("11111111-1111-1111-1111-111111111111")
		if !ok {
			http.Error(w, "game not found", http.StatusNotFound)
			return
		}

		_ = json.NewEncoder(w).Encode(session.Snapshot())
	})

	log.Printf("server started on :%s", cfg.HTTPPort)
	//log.Fatal(http.ListenAndServe(":"+cfg.HTTPPort, nil))
	log.Fatal(http.ListenAndServe(":"+cfg.HTTPPort, withCORS(http.DefaultServeMux)))
}
