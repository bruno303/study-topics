package kafka

import (
	"fmt"

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

	// Delivery report handler for produced messages
	go func() {
		for e := range producer.Events() {
			switch ev := e.(type) {
			case *libkafka.Message:
				if ev.TopicPartition.Error != nil {
					fmt.Printf("Delivery failed: %v\n", ev.TopicPartition)
				} else {
					fmt.Printf("Delivered message to %v\n", ev.TopicPartition)
				}
			}
		}
	}()

	return Producer{producer: producer}, nil
}

func (p Producer) Produce(msg string, topic string) error {
	return p.producer.Produce(&libkafka.Message{
		TopicPartition: libkafka.TopicPartition{Topic: &topic, Partition: libkafka.PartitionAny},
		Value:          []byte(msg),
	}, nil)
}
