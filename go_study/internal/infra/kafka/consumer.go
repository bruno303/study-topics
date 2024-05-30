package kafka

import (
	"context"
	"fmt"
	"main/internal/config"
	"main/internal/infra/observability/trace"
	"main/internal/infra/utils/shutdown"
	"time"

	libkafka "github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

var tracerConsumer = trace.GetTracer("KafkaConsumer")

type MessageHandler interface {
	Process(ctx context.Context, msg string) error
}

type consumer struct {
	c       *libkafka.Consumer
	topic   string
	handler MessageHandler
}

func newConsumer(cfg config.KafkaConsumerConfigDetail, handler MessageHandler) (consumer, error) {
	kc, err := libkafka.NewConsumer(&libkafka.ConfigMap{
		"bootstrap.servers":     cfg.Host,
		"group.id":              cfg.GroupId,
		"auto.offset.reset":     "earliest",
		"broker.address.family": "v4",
		// "auto.commit.interval.ms": "100",
		"enable.auto.commit": "false",
	})
	if err != nil {
		return consumer{}, err
	}
	return consumer{
		c:       kc,
		topic:   cfg.Topic,
		handler: handler,
	}, nil
}

func (c consumer) Start() error {
	err := c.c.SubscribeTopics([]string{c.topic}, nil)
	if err != nil {
		return err
	}

	run := true
	shutdown.CreateListener(func() {
		fmt.Println("Stopping consumer")
		run = false
	})

	go func() {
		defer c.c.Close()
		for run {
			ctx := context.Background()

			ctx, span := tracerConsumer.StartSpan(ctx, "ReadMessage")
			defer span.End()

			msg, err := c.c.ReadMessage(time.Second)
			if err == nil {
				fmt.Printf("Message on %s: %s\n", msg.TopicPartition, string(msg.Value))
				go func() {
					c.handler.Process(ctx, string(msg.Value))
					fmt.Printf("Message processed: %s\n", msg.TopicPartition)
					go func() {
						c.c.CommitMessage(msg)
					}()
				}()
			} else if !err.(libkafka.Error).IsTimeout() {
				span.SetError(err)
				fmt.Printf("Consumer error: %v (%v)\n", err, msg)
			}
		}
	}()

	fmt.Printf("Consumer for topic %s started\n", c.topic)
	return nil
}
