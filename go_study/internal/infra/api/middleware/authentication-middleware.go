package middleware

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/bruno303/study-topics/go-study/internal/config"
)

type AuthMiddleware struct {
	BaseMiddleware
	enabled   bool
	secretKey string
}

type authError struct {
	Msg string
}

func (m *AuthMiddleware) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	if !m.enabled {
		m.Next(rw, r)
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		e := authError{
			Msg: "Authorization header is required",
		}
		b, _ := json.Marshal(e)
		rw.Header().Add("Content-Type", "application/json")
		rw.WriteHeader(http.StatusUnauthorized)
		rw.Write(b)
		return
	}

	parts := strings.Split(authHeader, " ")
	if parts[1] != m.secretKey {
		e := authError{
			Msg: "Invalid api key",
		}
		b, _ := json.Marshal(e)
		rw.Header().Add("Content-Type", "application/json")
		rw.WriteHeader(http.StatusUnauthorized)
		rw.Write(b)
		return
	}

	m.Next(rw, r)
}

func NewAuthMiddleware(cfg *config.Config) *AuthMiddleware {
	return &AuthMiddleware{
		enabled:   cfg.Application.Auth.Enabled,
		secretKey: cfg.Application.Auth.SecretKey,
	}
}
