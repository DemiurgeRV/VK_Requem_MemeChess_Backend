package ws

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"meme_chess/internal/game"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 8192
)

type Client struct {
	hub         *Hub
	gameService *game.Service
	conn        *websocket.Conn
	send        chan []byte
	userID      string
	gameIDs     map[string]struct{}
}

func NewClient(hub *Hub, gameService *game.Service, conn *websocket.Conn, userID string) *Client {
	return &Client{
		hub:         hub,
		gameService: gameService,
		conn:        conn,
		send:        make(chan []byte, 256),
		userID:      userID,
		gameIDs:     make(map[string]struct{}),
	}
}

func (c *Client) readPump() {
	defer func() {
		for gameID := range c.gameIDs {
			_ = c.gameService.LeaveGame(gameID, c.userID)
			c.broadcastGameState(gameID)
		}

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
		c.handleJoinGame(msg)

	case "game.move":
		c.handleGameMove(msg)

	default:
		c.sendError(msg.RequestID, "UNKNOWN_TYPE", "unknown message type")
	}
}

func (c *Client) handleJoinGame(msg IncomingMessage) {
	var payload JoinGamePayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		c.sendError(msg.RequestID, "BAD_REQUEST", "invalid join payload")
		return
	}

	if payload.GameID == "" {
		c.sendError(msg.RequestID, "BAD_REQUEST", "game_id is required")
		return
	}

	state, err := c.gameService.JoinGame(payload.GameID, c.userID)
	if err != nil {
		c.sendGameError(msg.RequestID, err)
		return
	}

	c.hub.joinRoom <- subscription{
		client: c,
		gameID: payload.GameID,
	}

	c.sendJSON(OutgoingMessage{
		Type:      "game.joined",
		RequestID: msg.RequestID,
		Payload:   state,
	})

	c.broadcastGameState(payload.GameID)
}

func (c *Client) handleGameMove(msg IncomingMessage) {
	var payload GameMovePayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		c.sendError(msg.RequestID, "BAD_REQUEST", "invalid move payload")
		return
	}

	if payload.GameID == "" {
		c.sendError(msg.RequestID, "BAD_REQUEST", "game_id is required")
		return
	}
	if payload.Move == "" {
		c.sendError(msg.RequestID, "BAD_REQUEST", "move is required")
		return
	}

	state, result, err := c.gameService.MakeMove(context.Background(), payload.GameID, c.userID, payload.Move)
	if err != nil {
		c.sendGameError(msg.RequestID, err)
		return
	}

	c.sendJSON(OutgoingMessage{
		Type:      "game.move.accepted",
		RequestID: msg.RequestID,
		Payload: map[string]string{
			"game_id": payload.GameID,
			"move":    payload.Move,
		},
	})

	c.broadcastJSON(payload.GameID, OutgoingMessage{
		Type:    "game.state",
		Payload: state,
	})

	if result.IsCapture {
		c.broadcastJSON(payload.GameID, OutgoingMessage{
			Type: "game.event.capture",
			Payload: map[string]string{
				"game_id":    payload.GameID,
				"by_user_id": c.userID,
				"move":       result.Move,
			},
		})
	}

	if result.IsCheck {
		c.broadcastJSON(payload.GameID, OutgoingMessage{
			Type: "game.event.check",
			Payload: map[string]string{
				"game_id":    payload.GameID,
				"by_user_id": c.userID,
				"move":       result.Move,
			},
		})
	}

	if result.IsCheckmate {
		c.broadcastJSON(payload.GameID, OutgoingMessage{
			Type: "game.event.checkmate",
			Payload: map[string]string{
				"game_id":    payload.GameID,
				"by_user_id": c.userID,
				"move":       result.Move,
			},
		})

		c.broadcastJSON(payload.GameID, OutgoingMessage{
			Type: "game.finished",
			Payload: map[string]string{
				"game_id":         payload.GameID,
				"winner_id":       state.WinnerID,
				"finished_reason": state.FinishedReason,
			},
		})
	}
}

func (c *Client) broadcastGameState(gameID string) {
	session, ok := c.gameService.GetSession(gameID)
	if !ok {
		return
	}

	c.broadcastJSON(gameID, OutgoingMessage{
		Type:    "game.state",
		Payload: session.Snapshot(),
	})
}

func (c *Client) broadcastJSON(gameID string, v interface{}) {
	data, err := json.Marshal(v)
	if err != nil {
		log.Println("marshal error:", err)
		return
	}

	c.hub.broadcast <- BroadcastMessage{
		GameID:  gameID,
		Payload: data,
	}
}

func (c *Client) sendGameError(requestID string, err error) {
	switch {
	case errors.Is(err, game.ErrGameNotFound):
		c.sendError(requestID, "GAME_NOT_FOUND", "game not found")
	case errors.Is(err, game.ErrForbidden):
		c.sendError(requestID, "FORBIDDEN", "you are not a participant of this game")
	case errors.Is(err, game.ErrGameFull):
		c.sendError(requestID, "GAME_FULL", "game already has two players")
	case errors.Is(err, game.ErrNotYourTurn):
		c.sendError(requestID, "NOT_YOUR_TURN", "it is not your turn")
	case errors.Is(err, game.ErrGameFinished):
		c.sendError(requestID, "GAME_FINISHED", "game already finished")
	case errors.Is(err, game.ErrGameNotActive):
		c.sendError(requestID, "GAME_NOT_ACTIVE", "game is not active yet")
	case errors.Is(err, game.ErrInvalidMove):
		c.sendError(requestID, "INVALID_MOVE", "invalid move")
	default:
		c.sendError(requestID, "INTERNAL_ERROR", "internal error")
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
