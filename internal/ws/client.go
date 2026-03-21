package ws

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 8192
)

type Client struct {
	hub     *Hub
	conn    *websocket.Conn
	send    chan []byte
	userID  string
	gameIDs map[string]struct{}
}

func NewClient(hub *Hub, conn *websocket.Conn, userID string) *Client {
	return &Client{
		hub:     hub,
		conn:    conn,
		send:    make(chan []byte, 256),
		userID:  userID,
		gameIDs: make(map[string]struct{}),
	}
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		_ = c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			break
		}

		var msg IncomingMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			c.sendError("", "BAD_REQUEST", "invalid message format")
			continue
		}

		c.handleIncomingMessage(msg)
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) handleIncomingMessage(msg IncomingMessage) {
	switch msg.Type {
	case "game.join":
		var payload JoinGamePayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			c.sendError(msg.RequestID, "BAD_REQUEST", "invalid join payload")
			return
		}

		if payload.GameID == "" {
			c.sendError(msg.RequestID, "BAD_REQUEST", "game_id is required")
			return
		}

		c.hub.joinRoom <- subscription{
			client: c,
			gameID: payload.GameID,
		}

		c.sendJSON(OutgoingMessage{
			Type:      "game.joined",
			RequestID: msg.RequestID,
			Payload: map[string]string{
				"game_id": payload.GameID,
				"user_id": c.userID,
			},
		})

	case "game.message":
		var payload GameMessagePayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			c.sendError(msg.RequestID, "BAD_REQUEST", "invalid game message payload")
			return
		}

		if payload.GameID == "" {
			c.sendError(msg.RequestID, "BAD_REQUEST", "game_id is required")
			return
		}

		out := OutgoingMessage{
			Type: "game.message",
			Payload: map[string]string{
				"game_id": payload.GameID,
				"user_id": c.userID,
				"message": payload.Message,
			},
		}

		data, err := json.Marshal(out)
		if err != nil {
			log.Println("marshal error:", err)
			c.sendError(msg.RequestID, "INTERNAL_ERROR", "failed to build response")
			return
		}

		c.hub.broadcast <- BroadcastMessage{
			GameID:  payload.GameID,
			Payload: data,
		}

	default:
		c.sendError(msg.RequestID, "UNKNOWN_TYPE", "unknown message type")
	}
}

func (c *Client) sendError(requestID, code, message string) {
	c.sendJSON(OutgoingMessage{
		Type:      "error",
		RequestID: requestID,
		Error: &ErrorBody{
			Code:    code,
			Message: message,
		},
	})
}

func (c *Client) sendJSON(v interface{}) {
	data, err := json.Marshal(v)
	if err != nil {
		return
	}

	select {
	case c.send <- data:
	default:
	}
}
