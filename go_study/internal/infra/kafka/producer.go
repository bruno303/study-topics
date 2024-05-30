package kafka

import (
	"context"
	"fmt"
	"main/internal/infra/observability/trace"
	"main/internal/infra/utils/shutdown"

	libkafka "github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

var tracerProducer = trace.GetTracer("Producer")

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
		fmt.Println("Stopping producer delivery report")
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
						fmt.Printf("Delivery failed: %v\n", ev.TopicPartition)
					} else {
						fmt.Printf("Delivered message to %v\n", ev.TopicPartition)
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
	_, span := tracerProducer.StartSpan(ctx, "Produce")
	defer span.End()
	return p.producer.Produce(&libkafka.Message{
		TopicPartition: libkafka.TopicPartition{Topic: &topic, Partition: libkafka.PartitionAny},
		Value:          []byte(msg),
	}, nil)
}
