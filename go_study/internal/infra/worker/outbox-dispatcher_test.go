package worker

import (
	"context"
	"errors"
	"testing"
	"time"

	applicationRepository "github.com/bruno303/study-topics/go-study/internal/application/repository"
	"github.com/bruno303/study-topics/go-study/internal/application/transaction"
	"github.com/bruno303/study-topics/go-study/internal/config"
	"go.uber.org/mock/gomock"
)

type fakeOutboxProducer struct {
	produceFn func(ctx context.Context, msg string, topic string) error
}

func (p fakeOutboxProducer) Produce(ctx context.Context, msg string, topic string) error {
	if p.produceFn != nil {
		return p.produceFn(ctx, msg, topic)
	}
	return nil
}

func (fakeOutboxProducer) Close() {}

func TestOutboxDispatcherWorker_DispatchOne_WhenPublishSucceeds_MarksMessagePublished(t *testing.T) {
	ctrl := gomock.NewController(t)
	transactionManager := transaction.NewMockTransactionManager(ctrl)
	uow := transaction.NewMockUnitOfWork(ctrl)
	outboxRepo := applicationRepository.NewMockOutboxRepository(ctrl)
	now := time.Now().Add(-time.Second)
	msg := &applicationRepository.OutboxMessage{
		ID:          "outbox-1",
		Topic:       "topic",
		Payload:     `{"id":"1"}`,
		Status:      applicationRepository.OutboxStatusPending,
		Attempts:    0,
		AvailableAt: now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	transactionManager.EXPECT().
		WithinTx(gomock.Any(), transaction.EmptyOpts(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, _ transaction.TransactionOpts, fn transaction.TransactionCallback) error {
			return fn(ctx, uow)
		}).Times(1)

	uow.EXPECT().OutboxRepository().Return(outboxRepo).Times(2)
	outboxRepo.EXPECT().ClaimNext(gomock.Any(), 5).Return(msg, nil).Times(1)
	outboxRepo.EXPECT().
		Update(gomock.Any(), gomock.AssignableToTypeOf(applicationRepository.OutboxMessage{})).
		DoAndReturn(func(_ context.Context, got applicationRepository.OutboxMessage) error {
			if got.Status != applicationRepository.OutboxStatusPublished {
				t.Fatalf("expected published status, got %s", got.Status)
			}
			if got.PublishedAt.IsZero() {
				t.Fatal("expected published at to be set")
			}
			if got.Attempts != 0 {
				t.Fatalf("expected attempts unchanged, got %d", got.Attempts)
			}
			return nil
		}).Times(1)

	subject := NewOutboxDispatcherWorker(transactionManager, fakeOutboxProducer{}, config.OutboxDispatcherConfig{
		Enabled:            true,
		PollIntervalMillis: time.Hour.Milliseconds(),
		BatchSize:          10,
		MaxAttempts:        5,
	})

	handled, err := subject.dispatchOne(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !handled {
		t.Fatal("expected message to be handled")
	}
}

func TestOutboxDispatcherWorker_DispatchOne_WhenPublishFails_RetiresMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	transactionManager := transaction.NewMockTransactionManager(ctrl)
	uow := transaction.NewMockUnitOfWork(ctrl)
	outboxRepo := applicationRepository.NewMockOutboxRepository(ctrl)
	expectedErr := errors.New("publish failed")
	now := time.Now().Add(-time.Second)
	msg := &applicationRepository.OutboxMessage{
		ID:          "outbox-2",
		Topic:       "topic",
		Payload:     `{"id":"1"}`,
		Status:      applicationRepository.OutboxStatusPending,
		Attempts:    0,
		AvailableAt: now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	transactionManager.EXPECT().
		WithinTx(gomock.Any(), transaction.EmptyOpts(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, _ transaction.TransactionOpts, fn transaction.TransactionCallback) error {
			return fn(ctx, uow)
		}).Times(1)

	uow.EXPECT().OutboxRepository().Return(outboxRepo).Times(2)
	outboxRepo.EXPECT().ClaimNext(gomock.Any(), 5).Return(msg, nil).Times(1)
	outboxRepo.EXPECT().
		Update(gomock.Any(), gomock.AssignableToTypeOf(applicationRepository.OutboxMessage{})).
		DoAndReturn(func(_ context.Context, got applicationRepository.OutboxMessage) error {
			if got.Status != applicationRepository.OutboxStatusPending {
				t.Fatalf("expected pending status, got %s", got.Status)
			}
			if got.Attempts != 1 {
				t.Fatalf("expected attempts incremented to 1, got %d", got.Attempts)
			}
			if got.AvailableAt.Before(time.Now()) {
				t.Fatal("expected next retry to be in the future")
			}
			if got.LastError != expectedErr.Error() {
				t.Fatalf("expected last error to be set, got %q", got.LastError)
			}
			return nil
		}).Times(1)

	subject := NewOutboxDispatcherWorker(transactionManager, fakeOutboxProducer{
		produceFn: func(context.Context, string, string) error {
			return expectedErr
		},
	}, config.OutboxDispatcherConfig{
		Enabled:            true,
		PollIntervalMillis: time.Hour.Milliseconds(),
		BatchSize:          10,
		MaxAttempts:        5,
	})

	handled, err := subject.dispatchOne(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !handled {
		t.Fatal("expected message to be handled")
	}
}

func TestOutboxDispatcherWorker_DispatchOne_WhenMaxAttemptsReached_MarksMessageFailed(t *testing.T) {
	ctrl := gomock.NewController(t)
	transactionManager := transaction.NewMockTransactionManager(ctrl)
	uow := transaction.NewMockUnitOfWork(ctrl)
	outboxRepo := applicationRepository.NewMockOutboxRepository(ctrl)
	expectedErr := errors.New("publish failed")
	now := time.Now().Add(-time.Second)
	msg := &applicationRepository.OutboxMessage{
		ID:          "outbox-3",
		Topic:       "topic",
		Payload:     `{"id":"1"}`,
		Status:      applicationRepository.OutboxStatusPending,
		Attempts:    4,
		AvailableAt: now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	start := time.Now()

	transactionManager.EXPECT().
		WithinTx(gomock.Any(), transaction.EmptyOpts(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, _ transaction.TransactionOpts, fn transaction.TransactionCallback) error {
			return fn(ctx, uow)
		}).Times(1)

	uow.EXPECT().OutboxRepository().Return(outboxRepo).Times(2)
	outboxRepo.EXPECT().ClaimNext(gomock.Any(), 5).Return(msg, nil).Times(1)
	outboxRepo.EXPECT().
		Update(gomock.Any(), gomock.AssignableToTypeOf(applicationRepository.OutboxMessage{})).
		DoAndReturn(func(_ context.Context, got applicationRepository.OutboxMessage) error {
			if got.Status != applicationRepository.OutboxStatusFailed {
				t.Fatalf("expected failed status, got %s", got.Status)
			}
			if got.Attempts != 5 {
				t.Fatalf("expected attempts incremented to 5, got %d", got.Attempts)
			}
			if got.AvailableAt.Before(start) {
				t.Fatal("expected failed row availability to be set now or later")
			}
			if got.LastError != expectedErr.Error() {
				t.Fatalf("expected last error to be set, got %q", got.LastError)
			}
			return nil
		}).Times(1)

	subject := NewOutboxDispatcherWorker(transactionManager, fakeOutboxProducer{
		produceFn: func(context.Context, string, string) error {
			return expectedErr
		},
	}, config.OutboxDispatcherConfig{
		Enabled:            true,
		PollIntervalMillis: time.Hour.Milliseconds(),
		BatchSize:          10,
		MaxAttempts:        5,
	})

	handled, err := subject.dispatchOne(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !handled {
		t.Fatal("expected message to be handled")
	}
}
