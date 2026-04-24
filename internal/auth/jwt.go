package auth

import (
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

type JWTManager struct {
	secret  []byte
	mu      sync.RWMutex
	revoked map[string]time.Time
}

func NewJWTManager(secret string) *JWTManager {
	return &JWTManager{
		secret:  []byte(secret),
		revoked: make(map[string]time.Time),
	}
}

func (m *JWTManager) Generate(userID string) (string, error) {
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

func (m *JWTManager) Parse(tokenString string) (*Claims, error) {
	return m.parse(tokenString, false)
}

func (m *JWTManager) Revoke(tokenString string) error {
	claims, err := m.parse(tokenString, true)
	if err != nil {
		return err
	}

	expiresAt := time.Now().Add(24 * time.Hour)
	if claims.ExpiresAt != nil && !claims.ExpiresAt.Time.IsZero() {
		expiresAt = claims.ExpiresAt.Time
	}

	tokenString = strings.TrimSpace(tokenString)

	m.mu.Lock()
	defer m.mu.Unlock()

	m.cleanupExpiredLocked()
	m.revoked[tokenString] = expiresAt

	return nil
}

func (m *JWTManager) parse(tokenString string, allowRevoked bool) (*Claims, error) {
	tokenString = strings.TrimSpace(tokenString)
	if tokenString == "" {
		return nil, errors.New("missing token")
	}
	if !allowRevoked && m.isRevoked(tokenString) {
		return nil, errors.New("token revoked")
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, errors.New("unexpected signing method")
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

func (m *JWTManager) ClaimsFromAuthorizationHeader(authorization string) (*Claims, error) {
	raw := strings.TrimSpace(authorization)
	prefix := "bearer "
	if len(raw) < len(prefix) || strings.ToLower(raw[:len(prefix)]) != prefix {
		return nil, errors.New("missing bearer token")
	}
	token := strings.TrimSpace(raw[len(prefix):])
	if token == "" {
		return nil, errors.New("missing bearer token")
	}
	return m.Parse(token)
}

func (m *JWTManager) isRevoked(token string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.cleanupExpiredLocked()
	expiresAt, ok := m.revoked[token]
	if !ok {
		return false
	}

	if time.Now().After(expiresAt) {
		delete(m.revoked, token)
		return false
	}

	return true
}

func (m *JWTManager) cleanupExpiredLocked() {
	now := time.Now()
	for token, expiresAt := range m.revoked {
		if now.After(expiresAt) {
			delete(m.revoked, token)
		}
	}
}
