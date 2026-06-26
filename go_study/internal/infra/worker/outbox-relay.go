package worker

import (
	"context"
	"encoding/json"
	"time"

	"github.com/bruno303/study-topics/go-study/internal/application/transaction"
	"github.com/bruno303/study-topics/go-study/internal/config"
	"github.com/bruno303/study-topics/go-study/internal/crosscutting/observability/log"
	"github.com/bruno303/study-topics/go-study/internal/crosscutting/observability/trace"
	"github.com/bruno303/study-topics/go-study/internal/infra/utils/shutdown"
)

type Producer interface {
	ProduceSync(ctx context.Context, msg string, topic string, headers map[string]string) error
}

type OutboxRelayWorker struct {
	txManager TransactionManager
	prod      Producer
	cfg       config.OutboxRelayConfig
}

func NewOutboxRelayWorker(txManager TransactionManager, prod Producer, cfg config.OutboxRelayConfig) OutboxRelayWorker {
	return OutboxRelayWorker{
		txManager: txManager,
		prod:      prod,
		cfg:       cfg,
	}
}

func (w OutboxRelayWorker) Start() {
	if !w.cfg.Enabled {
		log.Log().Info(context.Background(), "OutboxRelayWorker disabled")
		return
	}

	run := true
	nextTick := time.NewTicker(time.Duration(w.cfg.PollIntervalMs) * time.Millisecond)

	shutdown.CreateListener(func() {
		log.Log().Info(context.Background(), "Stopping OutboxRelayWorker")
		nextTick.Stop()
		run = false
	})

	go func() {
		for range nextTick.C {
			if !run {
				return
			}
			_ = w.processBatch(context.Background())
		}
	}()

	log.Log().Info(context.Background(), "OutboxRelayWorker started")
}

func (w OutboxRelayWorker) processBatch(ctx context.Context) error {
	ctx, end := trace.Trace(ctx, trace.NameConfig("OutboxRelayWorker", "processBatch"))
	defer end()

	return w.txManager.WithinTx(ctx, transaction.EmptyOpts(), func(ctx context.Context, uow transaction.UnitOfWork) error {
		repo := uow.OutboxRepository()
		messages, err := repo.FetchPendingBatch(ctx, w.cfg.BatchSize)
		if err != nil {
			return err
		}

		for _, msg := range messages {
			var headers map[string]string
			if msg.Headers != "" && msg.Headers != "{}" {
				if err := json.Unmarshal([]byte(msg.Headers), &headers); err != nil {
					_ = repo.MarkFailed(ctx, msg.ID, "invalid headers: "+err.Error(), w.cfg.MaxAttempts)
					continue
				}
			}

			err := w.prod.ProduceSync(ctx, msg.Payload, msg.Topic, headers)
			if err != nil {
				_ = repo.MarkFailed(ctx, msg.ID, err.Error(), w.cfg.MaxAttempts)
			} else {
				_ = repo.MarkSent(ctx, msg.ID)
			}
		}
		return nil
	})
}
