package middleware

import (
	"context"

	"github.com/bruno303/study-topics/go-study/internal/infra/kafka/handlers"

	libkafka "github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type handlerMiddleware struct {
	BaseMiddleware
	handler handlers.MessageHandler
}

func NewMiddleware(h handlers.MessageHandler) *handlerMiddleware {
	return &handlerMiddleware{
		handler: h,
	}
}

func (m *handlerMiddleware) ProcessMessage(ctx context.Context, msg *libkafka.Message) {
	m.handler.Process(ctx, string(msg.Value))
}
