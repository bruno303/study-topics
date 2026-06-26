package worker

import (
	"context"
	"encoding/json"
	"math/rand"
	"strconv"
	"time"

	"github.com/bruno303/study-topics/go-study/internal/application/outbox"
	"github.com/bruno303/study-topics/go-study/internal/application/transaction"
	"github.com/bruno303/study-topics/go-study/internal/config"
	"github.com/bruno303/study-topics/go-study/internal/crosscutting/observability/log"
	"github.com/bruno303/study-topics/go-study/internal/crosscutting/observability/trace"
	"github.com/bruno303/study-topics/go-study/internal/crosscutting/observability/trace/attr"
	"github.com/bruno303/study-topics/go-study/internal/infra/utils/shutdown"

	"github.com/google/uuid"
)

//go:generate go tool mockgen -source=hello-producer.go -destination=mocks.go -package worker

type TransactionManager interface {
	WithinTx(context.Context, transaction.TransactionOpts, transaction.TransactionCallback) error
}

type HelloProducerWorker struct {
	txManager TransactionManager
	cfg       config.HelloProducerConfig
}

func NewHelloProducerWorker(txManager TransactionManager, cfg config.HelloProducerConfig) HelloProducerWorker {
	return HelloProducerWorker{
		txManager: txManager,
		cfg:       cfg,
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
		msgCount := 0
		for range nextTick.C {
			if msgCount >= w.cfg.MaxMessages {
				log.Log().Info(context.Background(), "Already sent the max quantity of messages: %d. Stopping...", w.cfg.MaxMessages)
				nextTick.Stop()
				run = false
			}
			if !run {
				return
			}
			_ = w.produceMessage(context.Background())
			msgCount++
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

	outboxMsg := &outbox.OutboxMessage{
		ID:             uuid.NewString(),
		Payload:        string(bytes),
		Topic:          w.cfg.Topic,
		Headers:        "{}",
		Status:         outbox.StatusPending,
		AttemptCounter: 0,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	return w.txManager.WithinTx(ctx, transaction.EmptyOpts(), func(ctx context.Context, uow transaction.UnitOfWork) error {
		return uow.OutboxRepository().Insert(ctx, outboxMsg)
	})
}
