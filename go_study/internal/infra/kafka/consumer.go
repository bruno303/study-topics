package kafka

import (
	"context"
	"fmt"
	"main/internal/config"
	"main/internal/crosscutting/observability"
	"main/internal/infra/utils/shutdown"
	"time"

	libkafka "github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type MessageHandler interface {
	Process(ctx context.Context, msg string) error
}

type consumer struct {
	c       *libkafka.Consumer
	topic   string
	handler MessageHandler
	cfg     config.KafkaConsumerConfigDetail
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
		cfg:     cfg,
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
			if c.cfg.TraceEnabled {
				observability.TraceWithAttr(ctx, "KafkaConsumer", "ReadMessage", func(ctx context.Context, sm observability.SpanModifier) {
					c.readMessage(ctx, sm)
				})
			} else {
				c.readMessage(ctx, observability.NoOpSpanModifier{})
			}
		}
	}()

	fmt.Printf("Consumer for topic %s started\n", c.topic)
	return nil
}

func (c consumer) readMessage(ctx context.Context, sm observability.SpanModifier) {
	msg, err := c.c.ReadMessage(time.Second)
	if err == nil {
		fmt.Printf("Message on %s: %s\n", msg.TopicPartition, string(msg.Value))
		go func() {
			observability.WithTracing(ctx, "KafkaConsumer", "ProcessMessage", func(ctx context.Context) {
				c.handler.Process(ctx, string(msg.Value))
				fmt.Printf("Message processed: %s\n", msg.TopicPartition)
				c.commitMessage(ctx, msg)
			})
		}()
	} else {
		sm.HandleError(err)
		if !err.(libkafka.Error).IsTimeout() {
			fmt.Printf("Consumer error: %v (%v)\n", err, msg)
		}
	}
}

func (c consumer) commitMessage(ctx context.Context, msg *libkafka.Message) {
	go func() {
		observability.WithTracing(
			ctx,
			"KafkaConsumer",
			"CommitMessage",
			func(ctx context.Context) {
				c.c.CommitMessage(msg)
			},
		)
	}()
}
