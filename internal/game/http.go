package game

import (
	"encoding/json"
	"net/http"

	"meme_chess/internal/auth"
)

type HTTP struct {
	Svc      *Service
	JWT      *auth.JWTManager
	JoinBase string // e.g. http://localhost:5173
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeAuthError(w http.ResponseWriter, err error) {
	msg := "unauthorized"
	if err != nil {
		msg = err.Error()
	}
	writeJSON(w, http.StatusUnauthorized, map[string]string{"error": msg})
}

// PostInvite creates a new room; host shares JoinBase/play/{game_id} with the opponent.
func (h *HTTP) PostInvite(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims, err := h.JWT.ClaimsFromAuthorizationHeader(r.Header.Get("Authorization"))
	if err != nil {
		writeAuthError(w, err)
		return
	}

	gameID, err := h.Svc.CreateInviteGame(r.Context(), claims.UserID, NewChessEngine())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create game"})
		return
	}

	joinURL := h.JoinBase + "/play/" + gameID
	writeJSON(w, http.StatusCreated, map[string]string{
		"game_id":  gameID,
		"join_url": joinURL,
	})
}

// guestGamePreview is returned to logged-in users who may take the second seat.
type guestGamePreview struct {
	GameID   string `json:"game_id"`
	Status   string `json:"status"`
	OpenSeat bool   `json:"open_seat"`
}

// GetGame returns full state for participants, or a short preview if the second seat is open.
func (h *HTTP) GetGame(w http.ResponseWriter, r *http.Request, gameID string) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims, err := h.JWT.ClaimsFromAuthorizationHeader(r.Header.Get("Authorization"))
	if err != nil {
		writeAuthError(w, err)
		return
	}

	session, ok := h.Svc.GetSession(gameID)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "game not found"})
		return
	}

	snap := session.Snapshot()
	if claims.UserID == snap.Player1ID || (snap.Player2ID != "" && claims.UserID == snap.Player2ID) {
		writeJSON(w, http.StatusOK, snap)
		return
	}

	if snap.Player2ID == "" && claims.UserID != snap.Player1ID {
		writeJSON(w, http.StatusOK, guestGamePreview{
			GameID:   snap.GameID,
			Status:   snap.Status,
			OpenSeat: true,
		})
		return
	}

	writeJSON(w, http.StatusForbidden, map[string]string{"error": "not a participant"})
}
