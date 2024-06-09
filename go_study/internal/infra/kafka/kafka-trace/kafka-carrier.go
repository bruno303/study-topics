package kafkatrace

import (
	"context"
	"main/internal/crosscutting/observability/trace"
	"main/pkg/utils/array"

	libkafka "github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"go.opentelemetry.io/otel"
)

type KafkaCarrier struct {
	msg *libkafka.Message
}

func NewKafkaCarrier(msg *libkafka.Message) KafkaCarrier {
	return KafkaCarrier{msg}
}

func Inject(ctx context.Context, msg *libkafka.Message) {
	carrier := NewKafkaCarrier(msg)
	otel.GetTextMapPropagator().Inject(ctx, carrier)
}

func Extract(ctx context.Context, msg *libkafka.Message, cfg *trace.TraceConfig) (context.Context, trace.EndFunc) {
	carrier := NewKafkaCarrier(msg)
	ctx = otel.GetTextMapPropagator().Extract(ctx, carrier)
	// always start a new trace with extracted one as parent (or root if parent is nil)
	return trace.Trace(ctx, cfg)
}

func (kc KafkaCarrier) Get(key string) string {
	header, ok := array.FirstOrNil(kc.msg.Headers, func(h libkafka.Header) bool {
		return h.Key == key
	})
	if !ok {
		return ""
	}
	return string(header.Value)
}

func (kc KafkaCarrier) Set(key string, value string) {
	for _, v := range kc.msg.Headers {
		if v.Key == key {
			v.Value = []byte(value)
			return
		}
	}
	kc.msg.Headers = append(kc.msg.Headers, libkafka.Header{Key: key, Value: []byte(value)})
}

func (kc KafkaCarrier) Keys() []string {
	return array.Map(kc.msg.Headers, func(h libkafka.Header) string {
		return h.Key
	})
}
