package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"main/internal/config"
	"main/internal/infra/observability/trace"
	"main/internal/infra/utils/shutdown"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

var tracer = trace.GetTracer("HelloProducerWorker")

type Producer interface {
	Produce(ctx context.Context, msg string, topic string) error
	Close()
}

type HelloProducerWorker struct {
	producer Producer
	cfg      config.HelloProducerConfig
}

func NewHelloProducerWorker(producer Producer, cfg config.HelloProducerConfig) HelloProducerWorker {
	return HelloProducerWorker{
		producer: producer,
		cfg:      cfg,
	}
}

type helloKafkaMsg struct {
	Id  string `json:"id"`
	Age int    `json:"age"`
}

func (w HelloProducerWorker) Start() {
	run := true
	nextTick := time.NewTicker(time.Duration(w.cfg.IntervalMillis) * time.Millisecond)

	shutdown.CreateListener(func() {
		fmt.Println("Stopping producer")
		nextTick.Stop()
		run = false
	})

	go func() {
		for range nextTick.C {
			if !run {
				return
			}
			_ = w.produceMessage(context.Background())
		}
	}()

	fmt.Println("HelloProducerWorker started")
}

func (w HelloProducerWorker) produceMessage(ctx context.Context) error {
	ctx, span := tracer.StartSpan(ctx, "produceMessage")
	defer span.End()
	msg := helloKafkaMsg{
		Id:  uuid.NewString(),
		Age: rand.Intn(150),
	}
	span.SetAttributes(trace.Attribute("id", msg.Id))
	bytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return w.producer.Produce(ctx, string(bytes), w.cfg.Topic)
}
