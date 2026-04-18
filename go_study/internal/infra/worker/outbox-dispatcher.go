package worker

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"time"

	applicationRepository "github.com/bruno303/study-topics/go-study/internal/application/repository"
	"github.com/bruno303/study-topics/go-study/internal/application/transaction"
	"github.com/bruno303/study-topics/go-study/internal/config"
	"github.com/bruno303/study-topics/go-study/internal/crosscutting/observability/log"
	"github.com/bruno303/study-topics/go-study/internal/crosscutting/observability/trace"
	"github.com/bruno303/study-topics/go-study/internal/crosscutting/observability/trace/attr"
	"github.com/bruno303/study-topics/go-study/internal/infra/utils/shutdown"
)

type OutboxDispatcherWorker struct {
	transactionManager transaction.TransactionManager
	producer           Producer
	cfg                config.OutboxDispatcherConfig
}

func NewOutboxDispatcherWorker(transactionManager transaction.TransactionManager, producer Producer, cfg config.OutboxDispatcherConfig) OutboxDispatcherWorker {
	cfg = applyDispatcherDefaults(cfg)
	return OutboxDispatcherWorker{
		transactionManager: transactionManager,
		producer:           producer,
		cfg:                cfg,
	}
}

func (w OutboxDispatcherWorker) Start() {
	if !w.cfg.Enabled {
		log.Log().Info(context.Background(), "OutboxDispatcherWorker disabled")
		return
	}

	nextTick := time.NewTicker(time.Duration(w.cfg.PollIntervalMillis) * time.Millisecond)
	stopCh := make(chan struct{})
	var stopOnce sync.Once

	shutdown.CreateListener(func() {
		log.Log().Info(context.Background(), "Stopping outbox dispatcher")
		nextTick.Stop()
		stopOnce.Do(func() {
			close(stopCh)
		})
	})

	go func() {
		for {
			select {
			case <-stopCh:
				return
			case <-nextTick.C:
				if err := w.dispatchBatch(context.Background()); err != nil {
					log.Log().Error(context.Background(), "outbox dispatch batch failed", err)
				}
			}
		}
	}()

	log.Log().Info(context.Background(), "OutboxDispatcherWorker started")
}

func (w OutboxDispatcherWorker) dispatchBatch(ctx context.Context) error {
	ctx, end := trace.Trace(ctx, trace.NameConfig("OutboxDispatcherWorker", "dispatchBatch"))
	defer end()

	maxBatch := w.cfg.BatchSize
	if maxBatch <= 0 {
		maxBatch = 10
	}

	for i := 0; i < maxBatch; i++ {
		handled, err := w.dispatchOne(ctx)
		if err != nil {
			return err
		}
		if !handled {
			return nil
		}
	}

	return nil
}

func (w OutboxDispatcherWorker) dispatchOne(ctx context.Context) (bool, error) {
	ctx, end := trace.Trace(ctx, trace.NameConfig("OutboxDispatcherWorker", "dispatchOne"))
	defer end()

	var handled bool
	var outcome string
	var messageID string
	var attempts int
	err := w.transactionManager.WithinTx(ctx, transaction.EmptyOpts(), func(txCtx context.Context, uow transaction.UnitOfWork) error {
		msg, err := uow.OutboxRepository().ClaimNext(txCtx, w.cfg.MaxAttempts)
		if err != nil {
			trace.InjectError(txCtx, err)
			return err
		}
		if msg == nil {
			return nil
		}
		handled = true
		messageID = msg.ID

		trace.InjectAttributes(txCtx,
			attr.New("outbox.id", msg.ID),
			attr.New("outbox.topic", msg.Topic),
			attr.New("outbox.attempts", strconv.Itoa(msg.Attempts)),
		)

		publishTimeout := w.cfg.PublishTimeout
		if publishTimeout <= 0 {
			publishTimeout = 10 * time.Second
		}

		publishCtx, cancel := context.WithTimeout(txCtx, publishTimeout)
		defer cancel()

		if err := w.producer.Produce(publishCtx, msg.Payload, msg.Topic); err != nil {
			outcome = "failed"
			attempts = msg.Attempts + 1
			return w.handleFailure(txCtx, uow, msg, err)
		}

		outcome = "published"
		return w.handleSuccess(txCtx, uow, msg)
	})
	if err != nil {
		return handled, err
	}

	if handled {
		switch outcome {
		case "published":
			log.Log().Info(ctx, "outbox message published: %s", messageID)
		case "failed":
			log.Log().Warn(ctx, "outbox message failed: %s attempt=%d", messageID, attempts)
		}
	}

	return handled, nil
}

func (w OutboxDispatcherWorker) handleSuccess(ctx context.Context, uow transaction.UnitOfWork, msg *applicationRepository.OutboxMessage) error {
	now := time.Now()
	msg.Status = applicationRepository.OutboxStatusPublished
	msg.PublishedAt = now
	msg.UpdatedAt = now
	msg.LastError = ""

	if err := uow.OutboxRepository().Update(ctx, *msg); err != nil {
		return fmt.Errorf("mark outbox message published: %w", err)
	}

	return nil
}

func (w OutboxDispatcherWorker) handleFailure(ctx context.Context, uow transaction.UnitOfWork, msg *applicationRepository.OutboxMessage, cause error) error {
	now := time.Now()
	msg.Attempts++
	msg.UpdatedAt = now
	msg.LastError = truncateError(cause.Error(), 1024)

	if msg.Attempts >= w.cfg.MaxAttempts {
		msg.Status = applicationRepository.OutboxStatusFailed
		msg.AvailableAt = now
	} else {
		msg.Status = applicationRepository.OutboxStatusPending
		msg.AvailableAt = nextRetryAt(msg.Attempts, now)
	}

	if err := uow.OutboxRepository().Update(ctx, *msg); err != nil {
		return fmt.Errorf("mark outbox message failed: %w", err)
	}

	return nil
}

func nextRetryAt(attempts int, now time.Time) time.Time {
	base := time.Second
	delay := base << max(attempts-1, 0)
	maxDelay := 60 * time.Second
	if delay > maxDelay {
		delay = maxDelay
	}
	jitter := time.Duration(rand.Int63n(int64(250 * time.Millisecond)))
	return now.Add(delay + jitter)
}

func truncateError(value string, maxLen int) string {
	if len(value) <= maxLen {
		return value
	}
	return value[:maxLen]
}

func applyDispatcherDefaults(cfg config.OutboxDispatcherConfig) config.OutboxDispatcherConfig {
	if cfg.PollIntervalMillis <= 0 {
		cfg.PollIntervalMillis = 5000
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 10
	}
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = applicationRepository.DefaultOutboxMaxAttempts
	}
	if cfg.PublishTimeout <= 0 {
		cfg.PublishTimeout = 10 * time.Second
	}
	return cfg
}
