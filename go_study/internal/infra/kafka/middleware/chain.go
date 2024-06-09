package middleware

import (
	"context"

	libkafka "github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type Chain struct {
	start Middleware
}

func NewChain(middlewares ...Middleware) Chain {
	for i := 0; i < len(middlewares)-1; i++ {
		middlewares[i].SetNext(middlewares[i+1])
	}
	return Chain{start: middlewares[0]}
}

func (c Chain) ProcessMessage(ctx context.Context, msg *libkafka.Message) {
	if c.start != nil {
		c.start.ProcessMessage(ctx, msg)
	}
}

type Middleware interface {
	ProcessMessage(ctx context.Context, msg *libkafka.Message)
	Next(ctx context.Context, msg *libkafka.Message)
	SetNext(m Middleware)
}

type BaseMiddleware struct {
	next Middleware
}

func (m *BaseMiddleware) Next(ctx context.Context, msg *libkafka.Message) {
	if m.next != nil {
		m.next.ProcessMessage(ctx, msg)
	}
}

func (m *BaseMiddleware) SetNext(mid Middleware) {
	m.next = mid
}
