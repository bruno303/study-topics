package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/bruno303/study-topics/go-study/internal/application/model"
	"github.com/bruno303/study-topics/go-study/internal/application/transaction"
	"github.com/bruno303/study-topics/go-study/internal/config"
	"github.com/bruno303/study-topics/go-study/internal/crosscutting/observability/log"
	"github.com/bruno303/study-topics/go-study/internal/crosscutting/observability/trace"
	"github.com/bruno303/study-topics/go-study/internal/crosscutting/observability/trace/attr"
	"github.com/bruno303/study-topics/go-study/internal/infra/utils/shutdown"
)

type OutboxSenderWorker struct {
	transactionManager transaction.TransactionManager
	producer           Producer
	cfg                config.OutboxSenderConfig
}

func NewOutboxSenderWorker(transactionManager transaction.TransactionManager, producer Producer, cfg config.OutboxSenderConfig) OutboxSenderWorker {
	return OutboxSenderWorker{
		transactionManager: transactionManager,
		producer:           producer,
		cfg:                cfg,
	}
}

func (w OutboxSenderWorker) Start() {
	ctx := context.Background()
	if !w.cfg.Enabled {
		log.Log().Info(context.Background(), "OutboxSenderWorker disabled")
		return
	}

	workerCtx, cancel := context.WithCancel(ctx)
	stopped := make(chan struct{})

	registered := shutdown.CreateListener(func() {
		log.Log().Info(context.Background(), "Stopping OutboxSenderWorker")
		cancel()
		<-stopped
	})
	if !registered {
		cancel()
		log.Log().Info(ctx, "OutboxSenderWorker start skipped during shutdown")
		return
	}

	ticker := time.NewTicker(time.Duration(w.cfg.IntervalMillis) * time.Millisecond)

	go func() {
		defer ticker.Stop()
		defer close(stopped)
		for {
			select {
			case <-workerCtx.Done():
				return
			case <-ticker.C:
				err := w.processTick(workerCtx)
				if err != nil {
					log.Log().Error(workerCtx, "Failed to process outbox tick", err)
				}
			}
		}
	}()

	log.Log().Info(ctx, "OutboxSenderWorker started")
}

func (w OutboxSenderWorker) processTick(ctx context.Context) error {
	ctx, end := trace.Trace(ctx, trace.NameConfig("OutboxSenderWorker", "processTick"))
	defer end()

	now := time.Now().UTC()
	messages, err := w.claimPending(ctx, now)
	if err != nil {
		trace.InjectError(ctx, err)
		return err
	}

	for _, message := range messages {
		err = w.publish(ctx, message)
		if err != nil {
			trace.InjectError(ctx, err)
		}
	}

	return nil
}

func (w OutboxSenderWorker) claimPending(ctx context.Context, now time.Time) ([]model.OutboxMessage, error) {
	var result []model.OutboxMessage

	err := w.transactionManager.WithinTx(ctx, transaction.EmptyOpts(), func(txCtx context.Context, uow transaction.UnitOfWork) error {
		messages, listErr := uow.OutboxRepository().ListPending(txCtx, w.cfg.BatchSize, w.cfg.MaxAttempts, now)
		if listErr != nil {
			return fmt.Errorf("list pending outbox messages: %w", listErr)
		}

		result = messages
		trace.InjectAttributes(txCtx, attr.New("outbox.claimed_count", fmt.Sprintf("%d", len(messages))))
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("claim outbox pending messages: %w", err)
	}

	return result, nil
}

func (w OutboxSenderWorker) publish(ctx context.Context, message model.OutboxMessage) error {
	ctx, end := trace.Trace(ctx, trace.NameConfig("OutboxSenderWorker", "publish"))
	defer end()

	trace.InjectAttributes(ctx,
		attr.New("outbox.id", message.Id),
		attr.New("outbox.topic", message.Topic),
		attr.New("outbox.attempt", fmt.Sprintf("%d", message.Attempt)),
	)

	err := w.producer.Produce(ctx, string(message.Payload), message.Topic, message.MessageKey, message.Headers)
	if err != nil {
		statusErr := w.markFailedPublish(ctx, message, err)
		if statusErr != nil {
			return fmt.Errorf("publish outbox message and mark failure: %w", statusErr)
		}
		return fmt.Errorf("publish outbox message: %w", err)
	}

	err = w.transactionManager.WithinTx(ctx, transaction.EmptyOpts(), func(txCtx context.Context, uow transaction.UnitOfWork) error {
		markErr := uow.OutboxRepository().MarkAsPublished(txCtx, message.Id, time.Now().UTC())
		if markErr != nil {
			return fmt.Errorf("mark outbox message as published: %w", markErr)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("update outbox message as published: %w", err)
	}

	return nil
}

func (w OutboxSenderWorker) markFailedPublish(ctx context.Context, message model.OutboxMessage, publishErr error) error {
	now := time.Now().UTC()
	nextAttempt := now.Add(time.Duration(w.cfg.RetryIntervalMillis) * time.Millisecond)
	if message.Attempt >= w.cfg.MaxAttempts {
		nextAttempt = time.Time{}
	}

	err := w.transactionManager.WithinTx(ctx, transaction.EmptyOpts(), func(txCtx context.Context, uow transaction.UnitOfWork) error {
		markErr := uow.OutboxRepository().MarkAsError(txCtx, message.Id, publishErr.Error(), nextAttempt)
		if markErr != nil {
			return fmt.Errorf("mark outbox message as error: %w", markErr)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("update outbox message as error: %w", err)
	}

	return nil
}
