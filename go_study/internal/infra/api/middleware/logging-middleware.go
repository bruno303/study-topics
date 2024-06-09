package middleware

import (
	"main/internal/crosscutting/observability/log"
	"net/http"
)

type LoggingMiddleware struct {
	BaseMiddleware
}

func (m *LoggingMiddleware) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	log.Log().Info(r.Context(), "Request received: %s %s", r.Method, r.URL.Path)
	m.Next(rw, r)
	log.Log().Info(r.Context(), "Request finished: %s %s", r.Method, r.URL.Path)
}

func LogMiddleware() *LoggingMiddleware {
	return &LoggingMiddleware{}
}
