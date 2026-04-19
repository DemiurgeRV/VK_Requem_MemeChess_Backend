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

type GameStickerPayload struct {
	GameID    string `json:"game_id"`
	StickerID string `json:"sticker_id,omitempty"`
	Title     string `json:"title,omitempty"`
	AssetURL  string `json:"asset_url"`
	MediaType string `json:"media_type,omitempty"`
	ImageURL  string `json:"image_url,omitempty"`
	VideoURL  string `json:"video_url,omitempty"`
	SoundURL  string `json:"sound_url,omitempty"`
}
