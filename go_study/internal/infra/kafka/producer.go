package kafka

import (
	"context"

	"github.com/bruno303/study-topics/go-study/internal/crosscutting/observability/log"
	"github.com/bruno303/study-topics/go-study/internal/crosscutting/observability/trace"
	kafkatrace "github.com/bruno303/study-topics/go-study/internal/infra/kafka/kafka-trace"
	"github.com/bruno303/study-topics/go-study/internal/infra/utils/shutdown"

	libkafka "github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type Producer struct {
	producer *libkafka.Producer
}

func NewProducer(bootstrapServer string) (Producer, error) {
	producer, err := libkafka.NewProducer(&libkafka.ConfigMap{
		"bootstrap.servers":     bootstrapServer,
		"broker.address.family": "v4",
	})
	if err != nil {
		return Producer{}, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	shutdown.CreateListener(func() {
		log.Log().Info(context.Background(), "Stopping producer delivery report")
		cancel()
	})

	// Delivery report handler for produced messages
	go func() {
		for {
			select {
			case <-ctx.Done():
				producer.Close()
				return
			case e, ok := <-producer.Events():
				if !ok {
					return
				}
				switch ev := e.(type) {
				case *libkafka.Message:
					if ev.TopicPartition.Error != nil {
						log.Log().Info(context.Background(), "Delivery failed: %v", ev.TopicPartition)
					} else {
						log.Log().Info(context.Background(), "Delivered message to %v", ev.TopicPartition)
					}
				}
			}
		}
	}()

	return Producer{producer: producer}, nil
}

func (p Producer) Close() {
	p.producer.Close()
}

func (p Producer) Produce(ctx context.Context, msg string, topic string, key string, headers map[string]string) error {
	ctx, end := trace.Trace(ctx, producerTrace("Produce"))
	defer end()

	var kafkaKey []byte
	if key != "" {
		kafkaKey = []byte(key)
	}

	kafkaMsg := &libkafka.Message{
		TopicPartition: libkafka.TopicPartition{Topic: &topic, Partition: libkafka.PartitionAny},
		Key:            kafkaKey,
		Value:          []byte(msg),
		Headers:        toKafkaHeaders(headers),
	}

	kafkatrace.Inject(ctx, kafkaMsg)

	return p.producer.Produce(kafkaMsg, nil)
}

func toKafkaHeaders(headers map[string]string) []libkafka.Header {
	if len(headers) == 0 {
		return nil
	}

	kafkaHeaders := make([]libkafka.Header, 0, len(headers))
	for key, value := range headers {
		kafkaHeaders = append(kafkaHeaders, libkafka.Header{Key: key, Value: []byte(value)})
	}

	return kafkaHeaders
}

func producerTrace(spanName string) *trace.TraceConfig {
	return &trace.TraceConfig{
		TraceName: "KafkaProducer",
		SpanName:  spanName,
		Kind:      trace.TraceKindProducer,
	}
}
