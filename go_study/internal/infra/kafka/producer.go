package kafka

import (
	"context"
	"main/internal/crosscutting/observability/log"
	"main/internal/crosscutting/observability/trace"
	kafkatrace "main/internal/infra/kafka/kafka-trace"
	"main/internal/infra/utils/shutdown"

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

	stop := false
	shutdown.CreateListener(func() {
		log.Log().Info(context.Background(), "Stopping producer delivery report")
		stop = true
	})

	// Delivery report handler for produced messages
	go func() {
		for {
			select {
			case e := <-producer.Events():
				switch ev := e.(type) {
				case *libkafka.Message:
					if ev.TopicPartition.Error != nil {
						log.Log().Info(context.Background(), "Delivery failed: %v", ev.TopicPartition)
					} else {
						log.Log().Info(context.Background(), "Delivered message to %v", ev.TopicPartition)
					}
				}
			default:
				if stop {
					producer.Close()
					return
				}
			}
		}
	}()

	return Producer{producer: producer}, nil
}

func (p Producer) Close() {
	p.producer.Close()
}

func (p Producer) Produce(ctx context.Context, msg string, topic string) error {
	ctx, end := trace.Trace(ctx, producerTrace("Produce"))
	defer end()

	kafkaMsg := &libkafka.Message{
		TopicPartition: libkafka.TopicPartition{Topic: &topic, Partition: libkafka.PartitionAny},
		Value:          []byte(msg),
	}

	kafkatrace.Inject(ctx, kafkaMsg)

	return p.producer.Produce(kafkaMsg, nil)
}

func producerTrace(spanName string) *trace.TraceConfig {
	return &trace.TraceConfig{
		TraceName: "KafkaProducer",
		SpanName:  spanName,
		Kind:      trace.TraceKindProducer,
	}
}
