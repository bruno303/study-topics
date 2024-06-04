package api

import (
	"errors"
	"net/http"
)

type Chain struct {
	start *Middleware
}

func NewChain(h MiddlewareFunc) (*Chain, error) {
	if h == nil {
		return nil, errors.New("chain must not be empty")
	}
	return &Chain{start: NewMiddleware(h)}, nil
}

func (c *Chain) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	if c.start != nil {
		c.start.ServeHTTP(rw, r, nil)
	}
}
