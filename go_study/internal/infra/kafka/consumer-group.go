package kafka

import (
	"github.com/bruno303/study-topics/go-study/internal/config"
	"github.com/bruno303/study-topics/go-study/internal/infra/kafka/handlers"
)

type ConsumerGroup struct {
	Consumers []consumer
	topic     string
}

func NewConsumerGroup(cfg config.KafkaConsumerConfigDetail, handler handlers.MessageHandler) (ConsumerGroup, error) {
	consumers := make([]consumer, 0, cfg.QntConsumers)
	for i := 0; i < cfg.QntConsumers; i++ {
		c, err := newConsumer(cfg, handler, i)
		if err != nil {
			return ConsumerGroup{}, err
		}
		consumers = append(consumers, c)
	}

	return ConsumerGroup{
		Consumers: consumers,
		topic:     cfg.Topic,
	}, nil
}

func (cg ConsumerGroup) Start() error {
	for _, c := range cg.Consumers {
		if err := c.Start(); err != nil {
			return err
		}
	}
	return nil
}
