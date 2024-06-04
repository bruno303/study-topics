package api

import (
	"context"
	"main/internal/crosscutting/observability/log"
	"net/http"

	"github.com/google/uuid"
)

type RequestIdContextKey struct {
	Key string
}

type Middleware struct {
	handler MiddlewareFunc
}

type MiddlewareFunc func(rw http.ResponseWriter, r *http.Request, next *Middleware)

func (m *Middleware) ServeHTTP(rw http.ResponseWriter, r *http.Request, next *Middleware) {
	m.handler(rw, r, next)
}

func NewMiddleware(h MiddlewareFunc) *Middleware {
	return &Middleware{handler: h}
}

func LogMiddleware(f MiddlewareFunc) MiddlewareFunc {
	return func(rw http.ResponseWriter, r *http.Request, next *Middleware) {
		log.Log().Info(r.Context(), "Request received: %s %s", r.Method, r.URL.Path)
		f(rw, r, next)
		log.Log().Info(r.Context(), "Request finished: %s %s", r.Method, r.URL.Path)
	}
}

func RequestIdMiddleware(f MiddlewareFunc) MiddlewareFunc {
	return func(rw http.ResponseWriter, r *http.Request, next *Middleware) {
		requestId := uuid.NewString()
		log.Log().Info(r.Context(), "RequestId generated: %s", requestId)
		ctx := context.WithValue(r.Context(), RequestIdContextKey{Key: "RequestIdContextKey"}, requestId)
		f(rw, r.WithContext(ctx), next)
	}
}

func ToMiddlewareFunc(f http.Handler) MiddlewareFunc {
	return func(rw http.ResponseWriter, r *http.Request, next *Middleware) {
		f.ServeHTTP(rw, r)
	}
}
