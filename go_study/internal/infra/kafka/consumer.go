package kafka

import (
	"context"
	"main/internal/config"
	"main/internal/crosscutting/observability/log"
	"main/internal/crosscutting/observability/trace"
	"main/internal/crosscutting/observability/trace/attr"
	"main/internal/infra/kafka/handlers"
	"main/internal/infra/kafka/middleware"
	"main/internal/infra/utils/shutdown"
	"strconv"
	"time"

	libkafka "github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type consumer struct {
	c       *libkafka.Consumer
	handler handlers.MessageHandler
	cfg     config.KafkaConsumerConfigDetail
	chain   middleware.Chain
}

func newConsumer(cfg config.KafkaConsumerConfigDetail, handler handlers.MessageHandler) (consumer, error) {
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
		handler: handler,
		cfg:     cfg,
		chain:   middleware.NewChain(middleware.NewLoggingMiddleware(), middleware.NewTracingMiddleware(), middleware.NewMiddleware(handler)),
	}, nil
}

func (c consumer) Start() error {
	if !c.cfg.Enabled {
		log.Log().Info(context.Background(), "Consumer for topic %s disabled", c.cfg.Topic)
		return nil
	}
	err := c.c.SubscribeTopics([]string{c.cfg.Topic}, nil)
	if err != nil {
		return err
	}

	run := true
	shutdown.CreateListener(func() {
		log.Log().Info(context.Background(), "Stopping consumer")
		run = false
	})

	go func() {
		defer c.c.Close()
		for run {
			c.readMessage(context.Background())
		}
		log.Log().Info(context.Background(), "Closing kafka consumer")
	}()

	log.Log().Info(context.Background(), "Consumer for topic %s started", c.cfg.Topic)
	return nil
}

func (c consumer) readMessage(ctx context.Context) {
	if c.cfg.TraceEnabled {
		_, end := trace.Trace(ctx, consumerTrace("ReadMessage"))
		defer end()

		trace.InjectAttributes(
			ctx,
			attr.New("kafka.bootstrapServer", c.cfg.Host),
			attr.New("kafka.topic", c.cfg.Topic),
			attr.New("kafka.consumerGroup", c.cfg.GroupId),
			attr.New("kafka.traceEnabled", strconv.FormatBool(c.cfg.TraceEnabled)),
		)
	}

	msg, err := c.c.ReadMessage(time.Second)
	if err == nil {
		log.Log().Info(ctx, "Message on %s: %s", msg.TopicPartition, string(msg.Value))
		go c.processMessage(ctx, msg)
	} else {
		// trace.InjectError(ctx, err)
		if !err.(libkafka.Error).IsTimeout() {
			log.Log().Error(ctx, "Consumer error", err)
		}
	}
}

func (c consumer) processMessage(ctx context.Context, msg *libkafka.Message) {
	c.chain.ProcessMessage(ctx, msg)
	log.Log().Info(ctx, "Message processed: %s", msg.TopicPartition)
	c.commitMessage(ctx, msg)
}

func (c consumer) commitMessage(ctx context.Context, msg *libkafka.Message) {
	go func() {
		_, end := trace.Trace(ctx, consumerTrace("CommitMessage"))
		defer end()
		trace.InjectAttributes(ctx, attr.New("kafka.message.key", string(msg.Key)))
		_, err := c.c.CommitMessage(msg)
		if err != nil {
			trace.InjectError(ctx, err)
		}
	}()
}

func consumerTrace(spanName string) *trace.TraceConfig {
	return &trace.TraceConfig{
		TraceName: "KafkaConsumer",
		SpanName:  spanName,
		Kind:      trace.TraceKindConsumer,
	}
}
