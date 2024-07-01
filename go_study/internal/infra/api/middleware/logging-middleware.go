package middleware

import (
	"net/http"

	"github.com/bruno303/study-topics/go-study/internal/crosscutting/observability/log"
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
