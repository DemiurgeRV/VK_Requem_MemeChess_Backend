package main

import (
	"encoding/json"
	"log"
	"net/http"

	"meme_chess/internal/auth"
	"meme_chess/internal/config"
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

	jwtManager := auth.NewJWTManager(cfg.JWTSecret)
	hub := ws.NewHub()
	wsHandler := ws.NewHandler(hub, jwtManager)

	go hub.Run()

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

	log.Printf("server started on :%s", cfg.HTTPPort)
	log.Fatal(http.ListenAndServe(":"+cfg.HTTPPort, withCORS(http.DefaultServeMux)))
}
