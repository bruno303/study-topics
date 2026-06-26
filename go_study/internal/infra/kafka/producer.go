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
					if isSyncMessage(ev) {
						continue
					}
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

func (p Producer) ProduceSync(ctx context.Context, msg string, topic string, headers map[string]string) error {
	ctx, end := trace.Trace(ctx, producerTrace("ProduceSync"))
	defer end()

	kafkaMsg := &libkafka.Message{
		TopicPartition: libkafka.TopicPartition{Topic: &topic, Partition: libkafka.PartitionAny},
		Value:          []byte(msg),
		Opaque:         "sync",
	}

	for k, v := range headers {
		kafkaMsg.Headers = append(kafkaMsg.Headers, libkafka.Header{Key: k, Value: []byte(v)})
	}

	kafkatrace.Inject(ctx, kafkaMsg)

	deliveryChan := make(chan libkafka.Event, 1)

	if err := p.producer.Produce(kafkaMsg, deliveryChan); err != nil {
		return err
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case e := <-deliveryChan:
		switch ev := e.(type) {
		case *libkafka.Message:
			if ev.TopicPartition.Error != nil {
				return ev.TopicPartition.Error
			}
			return nil
		default:
			return nil
		}
	}
}

func isSyncMessage(msg *libkafka.Message) bool {
	opaque, ok := msg.Opaque.(string)
	return ok && opaque == "sync"
}

func producerTrace(spanName string) *trace.TraceConfig {
	return &trace.TraceConfig{
		TraceName: "KafkaProducer",
		SpanName:  spanName,
		Kind:      trace.TraceKindProducer,
	}
}
