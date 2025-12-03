package middleware

import (
	"net/http"
	"strings"
)

type AdminMiddleware struct {
	apiKey string
}

func NewAdminMiddleware(apiKey string) AdminMiddleware {
	return AdminMiddleware{
		apiKey: apiKey,
	}
}

func (m *AdminMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization := r.Header.Get("Authorization")

		apiKey := strings.Replace(authorization, "Bearer ", "", 1)

		if apiKey != m.apiKey {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
