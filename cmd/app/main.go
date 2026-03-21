package main

import (
	"encoding/json"
	"log"
	"net/http"

	"meme_chess/internal/auth"
	"meme_chess/internal/config"
	"meme_chess/internal/ws"
)

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
	log.Fatal(http.ListenAndServe(":"+cfg.HTTPPort, nil))
}
