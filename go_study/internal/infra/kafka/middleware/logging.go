package middleware

import (
	"context"

	"github.com/bruno303/study-topics/go-study/internal/crosscutting/observability/log"

	libkafka "github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type LoggingMiddleware struct {
	BaseMiddleware
}

func NewLoggingMiddleware() *LoggingMiddleware {
	return &LoggingMiddleware{}
}

func (m *LoggingMiddleware) ProcessMessage(ctx context.Context, msg *libkafka.Message) {
	log.Log().Info(
		ctx,
		"Message received from topic %s, partition %d, offset %s, key %s",
		*msg.TopicPartition.Topic,
		msg.TopicPartition.Partition,
		msg.TopicPartition.Offset,
		string(msg.Key),
	)
	m.Next(ctx, msg)
}
