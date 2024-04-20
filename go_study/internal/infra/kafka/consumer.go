package kafka

import (
	"fmt"
	"time"

	libkafka "github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type MessageHandler interface {
	Process(msg string) error
}

type consumer struct {
	c       *libkafka.Consumer
	topic   string
	handler MessageHandler
}

func newConsumer(cfg *Config) (consumer, error) {
	kc, err := libkafka.NewConsumer(&libkafka.ConfigMap{
		"bootstrap.servers":     cfg.Host,
		"group.id":              cfg.GroupId,
		"auto.offset.reset":     "earliest",
		"broker.address.family": "v4",
	})
	if err != nil {
		return consumer{}, err
	}
	return consumer{
		c:       kc,
		topic:   cfg.Topic,
		handler: cfg.Handler,
	}, nil
}

func (c consumer) Start() error {
	err := c.c.SubscribeTopics([]string{c.topic}, nil)
	if err != nil {
		return err
	}

	go func() {
		run := true
		for run {
			msg, err := c.c.ReadMessage(time.Second)
			if err == nil {
				fmt.Printf("Message on %s: %s\n", msg.TopicPartition, string(msg.Value))
				c.handler.Process(string(msg.Value))
			} else if !err.(libkafka.Error).IsTimeout() {
				fmt.Printf("Consumer error: %v (%v)\n", err, msg)
			}
		}
		c.c.Close()
	}()

	fmt.Printf("Consumer for topic %s started\n", c.topic)
	return nil
}
