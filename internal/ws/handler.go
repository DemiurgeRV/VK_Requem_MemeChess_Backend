package ws

import (
	"net/http"

	"meme_chess/internal/auth"

	"github.com/gorilla/websocket"
)

type Handler struct {
	hub        *Hub
	jwtManager *auth.JWTManager
	upgrader   websocket.Upgrader
}

func NewHandler(hub *Hub, jwtManager *auth.JWTManager) *Handler {
	return &Handler{
		hub:        hub,
		jwtManager: jwtManager,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				origin := r.Header.Get("Origin")
				switch origin {
				case "http://localhost:3000":
					return true
				default:
					return false
				}
			},
		},
	}
}

func (h *Handler) ServeWS(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "missing token", http.StatusUnauthorized)
		return
	}

	claims, err := h.jwtManager.Parse(token)
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	client := NewClient(h.hub, conn, claims.UserID)
	h.hub.register <- client

	go client.writePump()
	go client.readPump()
}
