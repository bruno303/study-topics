package middleware

import (
	"context"
	"main/internal/crosscutting/observability/trace"
	"main/internal/crosscutting/observability/trace/attr"
	kafkatrace "main/internal/infra/kafka/kafka-trace"

	libkafka "github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type TracingMiddleware struct {
	BaseMiddleware
}

func NewTracingMiddleware() *TracingMiddleware {
	return &TracingMiddleware{}
}

func (m *TracingMiddleware) ProcessMessage(ctx context.Context, msg *libkafka.Message) {
	ctx, endCtxSpan := kafkatrace.Extract(ctx, msg, trace.NameConfig("TracingMiddleware", "ProcessMessage"))
	defer endCtxSpan()
	trace.InjectAttributes(ctx, attr.New("kafka.message.key", string(msg.Key)))
	m.Next(ctx, msg)
}
