package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"meme_chess/internal/user"
)

type Handlers struct {
	Service *Service
}

type registerRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type authResponse struct {
	Token string     `json:"token"`
	User  userPublic `json:"user"`
}

type currencyResponse struct {
	ShopFunds int64 `json:"shop_funds"`
	GameFunds int64 `json:"game_funds"`
}

type userPublic struct {
	ID        string     `json:"id"`
	Username  string     `json:"username"`
	Email     *string    `json:"email,omitempty"`
	AvatarURL *string    `json:"avatar_url,omitempty"`
	ShopFunds int64      `json:"shop_funds"`
	GameFunds int64      `json:"game_funds"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func (h *Handlers) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	token, u, err := h.Service.Register(r.Context(), RegisterInput{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidUsername):
			writeError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, ErrWeakPassword):
			writeError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, ErrDuplicateUser):
			writeError(w, http.StatusConflict, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "registration failed")
		}
		return
	}

	writeJSON(w, http.StatusCreated, authResponse{
		Token: token,
		User:  buildUserPublic(u),
	})
}

func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	token, u, err := h.Service.Login(r.Context(), LoginInput{
		Login:    req.Login,
		Password: req.Password,
	})
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			writeError(w, http.StatusUnauthorized, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "login failed")
		return
	}
	if u == nil {
		writeError(w, http.StatusInternalServerError, "login failed")
		return
	}

	writeJSON(w, http.StatusOK, authResponse{
		Token: token,
		User:  buildUserPublic(u),
	})
}

func (h *Handlers) Me(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	u, err := h.Service.UserFromBearer(r.Context(), r.Header.Get("Authorization"))
	if err != nil {
		switch {
		case errors.Is(err, ErrMissingToken):
			writeError(w, http.StatusUnauthorized, err.Error())
		case errors.Is(err, ErrInvalidToken):
			writeError(w, http.StatusUnauthorized, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "failed to load user")
		}
		return
	}

	writeJSON(w, http.StatusOK, buildUserPublic(u))
}

func (h *Handlers) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := h.Service.Logout(r.Header.Get("Authorization"))
	if err != nil {
		switch {
		case errors.Is(err, ErrMissingToken), errors.Is(err, ErrInvalidToken):
			writeError(w, http.StatusUnauthorized, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "logout failed")
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *Handlers) Currency(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	u, err := h.Service.UserFromBearer(r.Context(), r.Header.Get("Authorization"))
	if err != nil {
		switch {
		case errors.Is(err, ErrMissingToken):
			writeError(w, http.StatusUnauthorized, err.Error())
		case errors.Is(err, ErrInvalidToken):
			writeError(w, http.StatusUnauthorized, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "failed to load user currency")
		}
		return
	}

	writeJSON(w, http.StatusOK, currencyResponse{
		ShopFunds: u.ShopCurrency,
		GameFunds: u.GameCurrency,
	})
}

func buildUserPublic(u *user.User) userPublic {
	if u == nil {
		return userPublic{}
	}

	username := strings.TrimSpace(u.Username)

	return userPublic{
		ID:        u.ID,
		Username:  username,
		Email:     u.Email,
		AvatarURL: u.AvatarURL,
		ShopFunds: u.ShopCurrency,
		GameFunds: u.GameCurrency,
		CreatedAt: &u.CreatedAt,
	}
}
