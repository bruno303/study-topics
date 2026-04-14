package worker

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/bruno303/study-topics/go-study/internal/application/model"
	applicationRepository "github.com/bruno303/study-topics/go-study/internal/application/repository"
	"github.com/bruno303/study-topics/go-study/internal/application/transaction"
	"github.com/bruno303/study-topics/go-study/internal/config"
	"github.com/bruno303/study-topics/go-study/internal/infra/utils/shutdown"
	"go.uber.org/mock/gomock"
)

func setupOutboxSenderShutdownState(t *testing.T) {
	t.Helper()

	shutdown.ResetForTests()
	t.Cleanup(func() {
		shutdown.ResetForTests()
	})
}

func TestOutboxSenderWorker_ProcessTick_WhenClaimAndPublishSucceed_MarksAsPublished(t *testing.T) {
	ctrl := gomock.NewController(t)
	tm := transaction.NewMockTransactionManager(ctrl)
	uowClaim := transaction.NewMockUnitOfWork(ctrl)
	uowPublish := transaction.NewMockUnitOfWork(ctrl)
	outboxClaim := applicationRepository.NewMockOutboxRepository(ctrl)
	outboxPublish := applicationRepository.NewMockOutboxRepository(ctrl)
	producer := NewMockProducer(ctrl)

	message := model.OutboxMessage{
		Id:         "outbox-1",
		Topic:      "topic-1",
		MessageKey: "message-key-1",
		Payload:    []byte(`{"key":"value"}`),
		Headers: map[string]string{
			"x-tenant": "tenant-a",
		},
		Attempt: 1,
	}

	uowClaim.EXPECT().OutboxRepository().Return(outboxClaim)
	outboxClaim.EXPECT().ListPending(gomock.Any(), 10, 3, gomock.Any()).Return([]model.OutboxMessage{message}, nil)
	uowPublish.EXPECT().OutboxRepository().Return(outboxPublish)
	outboxPublish.EXPECT().MarkAsPublished(gomock.Any(), "outbox-1", gomock.Any()).Return(nil)
	producer.EXPECT().Produce(gomock.Any(), string(message.Payload), message.Topic, message.MessageKey, message.Headers).Return(nil)

	callCount := 0
	tm.EXPECT().WithinTx(gomock.Any(), transaction.EmptyOpts(), gomock.Any()).Times(2).DoAndReturn(func(ctx context.Context, _ transaction.TransactionOpts, fn transaction.TransactionCallback) error {
		callCount++
		if callCount == 1 {
			return fn(ctx, uowClaim)
		}
		return fn(ctx, uowPublish)
	})

	subject := NewOutboxSenderWorker(tm, producer, config.OutboxSenderConfig{BatchSize: 10, MaxAttempts: 3})

	err := subject.processTick(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestOutboxSenderWorker_ProcessTick_WhenPublishFailsBeforeMaxAttempts_MarksAsErrorWithRetry(t *testing.T) {
	ctrl := gomock.NewController(t)
	tm := transaction.NewMockTransactionManager(ctrl)
	uowClaim := transaction.NewMockUnitOfWork(ctrl)
	uowMarkError := transaction.NewMockUnitOfWork(ctrl)
	outboxClaim := applicationRepository.NewMockOutboxRepository(ctrl)
	outboxMarkError := applicationRepository.NewMockOutboxRepository(ctrl)
	producer := NewMockProducer(ctrl)

	message := model.OutboxMessage{
		Id:         "outbox-retry",
		Topic:      "topic-1",
		MessageKey: "message-key-retry",
		Payload:    []byte(`retry-me`),
		Headers: map[string]string{
			"x-source": "worker",
		},
		Attempt: 1,
	}

	uowClaim.EXPECT().OutboxRepository().Return(outboxClaim)
	outboxClaim.EXPECT().ListPending(gomock.Any(), 10, 3, gomock.Any()).Return([]model.OutboxMessage{message}, nil)
	producer.EXPECT().Produce(gomock.Any(), string(message.Payload), message.Topic, message.MessageKey, message.Headers).Return(errors.New("publish failed"))
	uowMarkError.EXPECT().OutboxRepository().Return(outboxMarkError)
	outboxMarkError.EXPECT().MarkAsError(gomock.Any(), "outbox-retry", "publish failed", gomock.Any()).DoAndReturn(func(_ context.Context, _ string, _ string, nextAttempt time.Time) error {
		if nextAttempt.IsZero() {
			t.Fatal("expected next attempt to be set for retryable publish failure")
		}
		return nil
	})

	callCount := 0
	tm.EXPECT().WithinTx(gomock.Any(), transaction.EmptyOpts(), gomock.Any()).Times(2).DoAndReturn(func(ctx context.Context, _ transaction.TransactionOpts, fn transaction.TransactionCallback) error {
		callCount++
		if callCount == 1 {
			return fn(ctx, uowClaim)
		}
		return fn(ctx, uowMarkError)
	})

	subject := NewOutboxSenderWorker(tm, producer, config.OutboxSenderConfig{BatchSize: 10, MaxAttempts: 3, RetryIntervalMillis: 1000})

	err := subject.processTick(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestOutboxSenderWorker_MarkFailedPublish_WhenMaxAttemptsReached_MarksAsFailed(t *testing.T) {
	ctrl := gomock.NewController(t)
	tm := transaction.NewMockTransactionManager(ctrl)
	uow := transaction.NewMockUnitOfWork(ctrl)
	outboxRepository := applicationRepository.NewMockOutboxRepository(ctrl)

	uow.EXPECT().OutboxRepository().Return(outboxRepository)
	outboxRepository.EXPECT().MarkAsError(gomock.Any(), "outbox-failed", "publish failed", time.Time{}).Return(nil)
	tm.EXPECT().WithinTx(gomock.Any(), transaction.EmptyOpts(), gomock.Any()).DoAndReturn(func(ctx context.Context, _ transaction.TransactionOpts, fn transaction.TransactionCallback) error {
		return fn(ctx, uow)
	})

	subject := NewOutboxSenderWorker(tm, nil, config.OutboxSenderConfig{MaxAttempts: 3, RetryIntervalMillis: 1000})

	err := subject.markFailedPublish(context.Background(), model.OutboxMessage{Id: "outbox-failed", Attempt: 3}, errors.New("publish failed"))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestOutboxSenderWorker_ProcessTick_WhenClaimFails_ReturnsWrappedError(t *testing.T) {
	ctrl := gomock.NewController(t)
	tm := transaction.NewMockTransactionManager(ctrl)
	tm.EXPECT().WithinTx(gomock.Any(), transaction.EmptyOpts(), gomock.Any()).Return(errors.New("claim failed"))

	subject := NewOutboxSenderWorker(tm, nil, config.OutboxSenderConfig{BatchSize: 10, MaxAttempts: 3})

	err := subject.processTick(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "claim outbox pending messages") {
		t.Fatalf("expected claim error wrapper, got %v", err)
	}
}

func TestOutboxSenderWorker_Start_WhenDisabled_DoesNotStartLifecycle(t *testing.T) {
	ctrl := gomock.NewController(t)
	tm := transaction.NewMockTransactionManager(ctrl)
	producer := NewMockProducer(ctrl)

	subject := NewOutboxSenderWorker(tm, producer, config.OutboxSenderConfig{
		IntervalMillis: 1,
		BatchSize:      1,
		Enabled:        false,
	})

	subject.Start()
}

func TestOutboxSenderWorker_ProcessTick_WhenContextCanceled_PropagatesContextToDependencies(t *testing.T) {
	ctrl := gomock.NewController(t)
	tm := transaction.NewMockTransactionManager(ctrl)
	uowClaim := transaction.NewMockUnitOfWork(ctrl)
	uowPublish := transaction.NewMockUnitOfWork(ctrl)
	outboxClaim := applicationRepository.NewMockOutboxRepository(ctrl)
	outboxPublish := applicationRepository.NewMockOutboxRepository(ctrl)
	producer := NewMockProducer(ctrl)

	message := model.OutboxMessage{
		Id:         "outbox-canceled",
		Topic:      "topic-1",
		MessageKey: "key-canceled",
		Payload:    []byte(`payload`),
		Headers: map[string]string{
			"x-request-id": "req-1",
		},
		Attempt: 1,
	}

	uowClaim.EXPECT().OutboxRepository().Return(outboxClaim)
	outboxClaim.EXPECT().ListPending(gomock.Any(), 10, 3, gomock.Any()).DoAndReturn(func(ctx context.Context, _ int, _ int, _ time.Time) ([]model.OutboxMessage, error) {
		if !errors.Is(ctx.Err(), context.Canceled) {
			t.Fatalf("expected canceled context, got %v", ctx.Err())
		}
		return []model.OutboxMessage{message}, nil
	})
	producer.EXPECT().
		Produce(gomock.Any(), string(message.Payload), message.Topic, message.MessageKey, message.Headers).
		DoAndReturn(func(ctx context.Context, _, _, _ string, _ map[string]string) error {
			if !errors.Is(ctx.Err(), context.Canceled) {
				t.Fatalf("expected canceled context, got %v", ctx.Err())
			}
			return nil
		})
	uowPublish.EXPECT().OutboxRepository().Return(outboxPublish)
	outboxPublish.EXPECT().MarkAsPublished(gomock.Any(), "outbox-canceled", gomock.Any()).DoAndReturn(func(ctx context.Context, _ string, _ time.Time) error {
		if !errors.Is(ctx.Err(), context.Canceled) {
			t.Fatalf("expected canceled context, got %v", ctx.Err())
		}
		return nil
	})

	callCount := 0
	tm.EXPECT().WithinTx(gomock.Any(), transaction.EmptyOpts(), gomock.Any()).Times(2).DoAndReturn(func(ctx context.Context, _ transaction.TransactionOpts, fn transaction.TransactionCallback) error {
		if !errors.Is(ctx.Err(), context.Canceled) {
			t.Fatalf("expected canceled context in tx manager, got %v", ctx.Err())
		}
		callCount++
		if callCount == 1 {
			return fn(ctx, uowClaim)
		}
		return fn(ctx, uowPublish)
	})

	subject := NewOutboxSenderWorker(tm, producer, config.OutboxSenderConfig{BatchSize: 10, MaxAttempts: 3})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := subject.processTick(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestOutboxSenderWorker_Start_WhenShutdownTriggered_WaitsForWorkerLoopExit(t *testing.T) {
	setupOutboxSenderShutdownState(t)

	ctrl := gomock.NewController(t)
	tm := transaction.NewMockTransactionManager(ctrl)
	uowClaim := transaction.NewMockUnitOfWork(ctrl)
	outboxClaim := applicationRepository.NewMockOutboxRepository(ctrl)
	producer := NewMockProducer(ctrl)

	started := make(chan struct{})
	canceled := make(chan struct{})
	release := make(chan struct{})
	returned := make(chan struct{})
	var startedOnce sync.Once
	var canceledOnce sync.Once
	var returnedOnce sync.Once

	uowClaim.EXPECT().OutboxRepository().Return(outboxClaim).AnyTimes()
	outboxClaim.EXPECT().ListPending(gomock.Any(), 1, 3, gomock.Any()).DoAndReturn(func(ctx context.Context, _ int, _ int, _ time.Time) ([]model.OutboxMessage, error) {
		startedOnce.Do(func() {
			close(started)
		})
		<-ctx.Done()
		canceledOnce.Do(func() {
			close(canceled)
			<-release
			returnedOnce.Do(func() {
				close(returned)
			})
		})
		return []model.OutboxMessage{}, nil
	}).AnyTimes()
	tm.EXPECT().WithinTx(gomock.Any(), transaction.EmptyOpts(), gomock.Any()).DoAndReturn(func(ctx context.Context, _ transaction.TransactionOpts, fn transaction.TransactionCallback) error {
		return fn(ctx, uowClaim)
	}).AnyTimes()

	subject := NewOutboxSenderWorker(tm, producer, config.OutboxSenderConfig{
		IntervalMillis: 100,
		BatchSize:      1,
		MaxAttempts:    3,
		Enabled:        true,
	})

	subject.Start()

	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("expected worker loop to start processing before shutdown")
	}

	shutdown.Trigger()

	awaited := make(chan struct{})
	go func() {
		shutdown.AwaitAll()
		close(awaited)
	}()

	select {
	case <-canceled:
	case <-time.After(2 * time.Second):
		t.Fatal("expected shutdown to cancel worker context")
	}

	select {
	case <-awaited:
		t.Fatal("expected shutdown to wait for worker loop exit")
	default:
	}

	close(release)

	select {
	case <-returned:
	case <-time.After(2 * time.Second):
		t.Fatal("expected in-flight worker tick to finish after release")
	}

	select {
	case <-awaited:
	case <-time.After(2 * time.Second):
		t.Fatal("expected shutdown listener to complete after worker loop exit")
	}
}

func TestOutboxSenderWorker_Start_WhenShutdownAlreadyInProgress_DoesNotLaunchWorker(t *testing.T) {
	setupOutboxSenderShutdownState(t)

	ctrl := gomock.NewController(t)
	tm := transaction.NewMockTransactionManager(ctrl)
	producer := NewMockProducer(ctrl)

	shutdown.Trigger()

	subject := NewOutboxSenderWorker(tm, producer, config.OutboxSenderConfig{
		IntervalMillis: 1,
		BatchSize:      1,
		MaxAttempts:    3,
		Enabled:        true,
	})

	subject.Start()

	shutdown.AwaitAll()
}
