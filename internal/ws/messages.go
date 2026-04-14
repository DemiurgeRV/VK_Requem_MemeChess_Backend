package ws

import "encoding/json"

type IncomingMessage struct {
	Type      string          `json:"type"`
	RequestID string          `json:"request_id,omitempty"`
	Payload   json.RawMessage `json:"payload"`
}

type OutgoingMessage struct {
	Type      string      `json:"type"`
	RequestID string      `json:"request_id,omitempty"`
	Payload   interface{} `json:"payload,omitempty"`
	Error     *ErrorBody  `json:"error,omitempty"`
}

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type JoinGamePayload struct {
	GameID string `json:"game_id"`
}

type GameMovePayload struct {
	GameID string `json:"game_id"`
	Move   string `json:"move"`
}

type GameEmotePayload struct {
	GameID   string `json:"game_id"`
	EmoteMP4 string `json:"emote_mp4"`
}
