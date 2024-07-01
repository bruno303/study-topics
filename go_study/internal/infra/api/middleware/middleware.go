package middleware

import (
	"net/http"
)

type Middleware interface {
	http.Handler
	Next(rw http.ResponseWriter, r *http.Request)
	SetNext(Middleware)
}

type BaseMiddleware struct {
	next Middleware
}

func (m *BaseMiddleware) Next(rw http.ResponseWriter, r *http.Request) {
	if m.next != nil {
		m.next.ServeHTTP(rw, r)
	}
}

func (m *BaseMiddleware) SetNext(md Middleware) {
	m.next = md
}

type DefaultMiddleware struct {
	BaseMiddleware
	handler http.Handler
}

func (m *DefaultMiddleware) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	m.handler.ServeHTTP(rw, r)
}

func NewMiddleware(h http.Handler) *DefaultMiddleware {
	return &DefaultMiddleware{
		handler: h,
	}
}
