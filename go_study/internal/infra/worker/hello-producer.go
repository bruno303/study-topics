package worker

import (
	"context"
	"encoding/json"
	"math/rand"
	"strconv"
	"time"

	"github.com/bruno303/study-topics/go-study/internal/config"
	"github.com/bruno303/study-topics/go-study/internal/crosscutting/observability/log"
	"github.com/bruno303/study-topics/go-study/internal/crosscutting/observability/trace"
	"github.com/bruno303/study-topics/go-study/internal/crosscutting/observability/trace/attr"
	"github.com/bruno303/study-topics/go-study/internal/infra/utils/shutdown"

	"github.com/google/uuid"
)

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
	if !w.cfg.Enabled {
		log.Log().Info(context.Background(), "HelloProducerWorker disabled")
		return
	}
	run := true
	nextTick := time.NewTicker(time.Duration(w.cfg.IntervalMillis) * time.Millisecond)

	shutdown.CreateListener(func() {
		log.Log().Info(context.Background(), "Stopping producer")
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

	log.Log().Info(context.Background(), "HelloProducerWorker started")
}

func (w HelloProducerWorker) produceMessage(ctx context.Context) error {
	ctx, end := trace.Trace(ctx, trace.NameConfig("HelloProducerWorker", "produceMessage"))
	defer end()

	msg := helloKafkaMsg{
		Id:  uuid.NewString(),
		Age: rand.Intn(150),
	}

	trace.InjectAttributes(ctx, attr.New("msg.id", msg.Id), attr.New("msg.age", strconv.Itoa(msg.Age)))

	bytes, err := json.Marshal(msg)
	if err != nil {
		trace.InjectError(ctx, err)
		return err
	}
	return w.producer.Produce(ctx, string(bytes), w.cfg.Topic)
}
