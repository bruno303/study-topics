package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/bruno303/study-topics/go-study/internal/application/model"
	"github.com/bruno303/study-topics/go-study/internal/application/transaction"
	"github.com/bruno303/study-topics/go-study/internal/config"
	"github.com/bruno303/study-topics/go-study/internal/crosscutting/observability/log"
	"github.com/bruno303/study-topics/go-study/internal/crosscutting/observability/trace"
	"github.com/bruno303/study-topics/go-study/internal/crosscutting/observability/trace/attr"
	"github.com/bruno303/study-topics/go-study/internal/infra/utils/shutdown"

	"github.com/google/uuid"
)

//go:generate go tool mockgen -source=hello-producer.go -destination=mocks.go -package worker

type Producer interface {
	Produce(ctx context.Context, msg string, topic string, key string, headers map[string]string) error
	Close()
}

type HelloProducerWorker struct {
	transactionManager transaction.TransactionManager
	cfg                config.HelloProducerConfig
}

func NewHelloProducerWorker(transactionManager transaction.TransactionManager, cfg config.HelloProducerConfig) HelloProducerWorker {
	return HelloProducerWorker{
		transactionManager: transactionManager,
		cfg:                cfg,
	}
}

type helloKafkaMsg struct {
	Id  string `json:"id"`
	Age int    `json:"age"`
}

func (w HelloProducerWorker) Start() {
	ctx := context.Background()
	if !w.cfg.Enabled {
		log.Log().Info(context.Background(), "HelloProducerWorker disabled")
		return
	}

	workerCtx, cancel := context.WithCancel(ctx)
	stopped := make(chan struct{})

	registered := shutdown.CreateListener(func() {
		log.Log().Info(context.Background(), "Stopping HelloProducerWorker")
		cancel()
		<-stopped
	})
	if !registered {
		cancel()
		log.Log().Info(ctx, "HelloProducerWorker start skipped during shutdown")
		return
	}

	nextTick := time.NewTicker(time.Duration(w.cfg.IntervalMillis) * time.Millisecond)

	go func() {
		defer nextTick.Stop()
		defer close(stopped)
		msgCount := 0
		for {
			select {
			case <-workerCtx.Done():
				return
			case <-nextTick.C:
				if msgCount >= w.cfg.MaxMessages {
					log.Log().Info(workerCtx, "Already sent the max quantity of messages: %d. Stopping...", w.cfg.MaxMessages)
					return
				}

				_ = w.produceMessage(workerCtx)
				msgCount++
			}
		}
	}()

	log.Log().Info(ctx, "HelloProducerWorker started")
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
	err = w.transactionManager.WithinTx(ctx, transaction.EmptyOpts(), func(txCtx context.Context, uow transaction.UnitOfWork) error {
		_, enqueueErr := uow.OutboxRepository().Enqueue(txCtx, &model.OutboxMessage{
			Id:         uuid.NewString(),
			Topic:      w.cfg.Topic,
			MessageKey: msg.Id,
			Payload:    bytes,
		})
		if enqueueErr != nil {
			return fmt.Errorf("enqueue hello outbox message: %w", enqueueErr)
		}

		return nil
	})
	if err != nil {
		trace.InjectError(ctx, err)
		return err
	}

	return nil
}
