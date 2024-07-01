package kafka

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/bruno303/study-topics/go-study/internal/config"
	"github.com/bruno303/study-topics/go-study/internal/crosscutting/observability/log"
	"github.com/bruno303/study-topics/go-study/internal/crosscutting/observability/trace"
	"github.com/bruno303/study-topics/go-study/internal/crosscutting/observability/trace/attr"
	"github.com/bruno303/study-topics/go-study/internal/infra/kafka/handlers"
	"github.com/bruno303/study-topics/go-study/internal/infra/kafka/middleware"
	"github.com/bruno303/study-topics/go-study/internal/infra/utils/shutdown"

	libkafka "github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type consumer struct {
	c          *libkafka.Consumer
	handler    handlers.MessageHandler
	cfg        config.KafkaConsumerConfigDetail
	chain      middleware.Chain
	identifier int
}

func newConsumer(cfg config.KafkaConsumerConfigDetail, handler handlers.MessageHandler, identifier int) (consumer, error) {
	kc, err := libkafka.NewConsumer(&libkafka.ConfigMap{
		"bootstrap.servers":       cfg.Host,
		"group.id":                cfg.GroupId,
		"auto.offset.reset":       cfg.OffsetReset,
		"broker.address.family":   "v4",
		"auto.commit.interval.ms": strconv.FormatInt(cfg.AutoCommitInterval.Milliseconds(), 10),
		"enable.auto.commit":      strconv.FormatBool(cfg.AutoCommit),
	})
	if err != nil {
		return consumer{}, err
	}
	return consumer{
		c:          kc,
		handler:    handler,
		cfg:        cfg,
		chain:      middleware.NewChain(middleware.NewLoggingMiddleware(), middleware.NewTracingMiddleware(), middleware.NewMiddleware(handler)),
		identifier: identifier,
	}, nil
}

func (c consumer) Start() error {
	if !c.cfg.Enabled {
		log.Log().Info(context.Background(), "[%v] Consumer for topic %s disabled", c.identifier, c.cfg.Topic)
		return nil
	}
	err := c.c.SubscribeTopics([]string{c.cfg.Topic}, nil)
	if err != nil {
		return err
	}

	run := true
	closeChan := make(chan struct{})

	shutdown.CreateListener(func() {
		log.Log().Info(context.Background(), "[%v] Stopping consumer", c.identifier)
		run = false
		<-closeChan
		log.Log().Info(context.Background(), "[%v] Closing kafka consumer", c.identifier)
		c.c.Close()
	})

	go func() {
		for run {
			c.readMessage(context.Background())
		}
		closeChan <- struct{}{}
	}()

	log.Log().Info(context.Background(), "[%v] Consumer for topic %s started", c.identifier, c.cfg.Topic)
	return nil
}

func (c consumer) readMessage(ctx context.Context) {
	log.Log().Debug(ctx, "[%v] Reading kafka message", c.identifier)
	if c.cfg.TraceEnabled {
		ctx, end := trace.Trace(ctx, consumerTrace("ReadMessage"))
		defer end()

		trace.InjectAttributes(
			ctx,
			attr.New("kafka.bootstrapServer", c.cfg.Host),
			attr.New("kafka.topic", c.cfg.Topic),
			attr.New("kafka.consumerGroup", c.cfg.GroupId),
			attr.New("kafka.traceEnabled", strconv.FormatBool(c.cfg.TraceEnabled)),
			attr.New("kafka.consumer.identifier", strconv.Itoa(c.identifier)),
		)
	}

	msg, err := c.c.ReadMessage(time.Second)
	if err == nil {
		log.Log().Info(ctx, "[%v] Message on %s: %s", c.identifier, msg.TopicPartition, string(msg.Value))
		c.processMessage(ctx, msg)
	} else {
		trace.InjectError(ctx, err)
		if !err.(libkafka.Error).IsTimeout() {
			log.Log().Error(ctx, fmt.Sprintf("[%v] Consumer error", c.identifier), err)
		}
	}
}

func (c consumer) processMessage(ctx context.Context, msg *libkafka.Message) {
	c.chain.ProcessMessage(ctx, msg)
	log.Log().Info(ctx, "Message processed: %s", msg.TopicPartition)
	c.commitMessage(ctx, msg)
}

func (c consumer) commitMessage(ctx context.Context, msg *libkafka.Message) {
	if c.cfg.AutoCommit {
		log.Log().Debug(ctx, "[%v] Auto commit enabled, skipping manual commit", c.identifier)
		return
	}
	if c.cfg.AsyncCommit {
		log.Log().Debug(ctx, "[%v] Committing message async", c.identifier)
		go c.doCommit(ctx, msg)
	} else {
		log.Log().Debug(ctx, "[%v] Committing message sync", c.identifier)
		c.doCommit(ctx, msg)
	}
}

func (c consumer) doCommit(ctx context.Context, msg *libkafka.Message) {
	ctx, end := trace.Trace(ctx, consumerTrace("CommitMessage"))
	defer end()
	trace.InjectAttributes(
		ctx,
		attr.New("kafka.message.key", string(msg.Key)),
		attr.New("kafka.consumer.identifier", strconv.Itoa(c.identifier)),
	)
	_, err := c.c.CommitMessage(msg)
	if err != nil {
		trace.InjectError(ctx, err)
	}
}

func consumerTrace(spanName string) *trace.TraceConfig {
	return &trace.TraceConfig{
		TraceName: "KafkaConsumer",
		SpanName:  spanName,
		Kind:      trace.TraceKindConsumer,
	}
}
