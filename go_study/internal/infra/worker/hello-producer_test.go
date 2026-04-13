package worker

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/bruno303/study-topics/go-study/internal/application/model"
	applicationRepository "github.com/bruno303/study-topics/go-study/internal/application/repository"
	"github.com/bruno303/study-topics/go-study/internal/application/transaction"
	"github.com/bruno303/study-topics/go-study/internal/config"
	"go.uber.org/mock/gomock"
)

func TestProduceMessage_WhenTransactionAndEnqueueSucceed_ReturnsNil(t *testing.T) {
	ctrl := gomock.NewController(t)
	tm := transaction.NewMockTransactionManager(ctrl)
	uow := transaction.NewMockUnitOfWork(ctrl)
	outboxRepository := applicationRepository.NewMockOutboxRepository(ctrl)

	uow.EXPECT().OutboxRepository().Return(outboxRepository)
	outboxRepository.EXPECT().Enqueue(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, message *model.OutboxMessage) (*model.OutboxMessage, error) {
		if message == nil {
			t.Fatal("expected outbox message to be provided")
		}
		if message.Topic != "topic" {
			t.Fatalf("expected topic %q, got %q", "topic", message.Topic)
		}
		if message.MessageKey == "" {
			t.Fatal("expected message key to be generated")
		}
		if len(message.Payload) == 0 {
			t.Fatal("expected payload bytes to be generated")
		}
		return message, nil
	})

	tm.EXPECT().WithinTx(gomock.Any(), transaction.EmptyOpts(), gomock.Any()).DoAndReturn(func(ctx context.Context, _ transaction.TransactionOpts, fn transaction.TransactionCallback) error {
		return fn(ctx, uow)
	})

	subject := NewHelloProducerWorker(tm, config.HelloProducerConfig{
		IntervalMillis: time.Hour.Milliseconds(),
		Topic:          "topic",
		Enabled:        true,
	})

	if err := subject.produceMessage(context.Background()); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestProduceMessage_WhenEnqueueFails_ReturnsWrappedError(t *testing.T) {
	ctrl := gomock.NewController(t)
	tm := transaction.NewMockTransactionManager(ctrl)
	uow := transaction.NewMockUnitOfWork(ctrl)
	outboxRepository := applicationRepository.NewMockOutboxRepository(ctrl)

	uow.EXPECT().OutboxRepository().Return(outboxRepository)
	outboxRepository.EXPECT().Enqueue(gomock.Any(), gomock.Any()).Return(nil, errors.New("enqueue failed"))
	tm.EXPECT().WithinTx(gomock.Any(), transaction.EmptyOpts(), gomock.Any()).DoAndReturn(func(ctx context.Context, _ transaction.TransactionOpts, fn transaction.TransactionCallback) error {
		return fn(ctx, uow)
	})

	subject := NewHelloProducerWorker(tm, config.HelloProducerConfig{Topic: "topic"})

	err := subject.produceMessage(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "enqueue hello outbox message") {
		t.Fatalf("expected wrapped enqueue error, got %v", err)
	}
}

func TestProduceMessage_WhenWithinTxFails_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	tm := transaction.NewMockTransactionManager(ctrl)
	tm.EXPECT().WithinTx(gomock.Any(), transaction.EmptyOpts(), gomock.Any()).Return(errors.New("tx failed"))

	subject := NewHelloProducerWorker(tm, config.HelloProducerConfig{Topic: "topic"})

	err := subject.produceMessage(context.Background())
	if err == nil || err.Error() != "tx failed" {
		t.Fatalf("expected tx error, got %v", err)
	}
}

func TestHelloProducerWorker_Start_WhenDisabled_DoesNotStartLifecycle(t *testing.T) {
	ctrl := gomock.NewController(t)
	tm := transaction.NewMockTransactionManager(ctrl)

	subject := NewHelloProducerWorker(tm, config.HelloProducerConfig{
		IntervalMillis: 1,
		Enabled:        false,
		MaxMessages:    1,
		Topic:          "topic",
	})

	subject.Start()
}

func TestProduceMessage_WhenContextCanceled_PropagatesContextToTransactionAndRepository(t *testing.T) {
	ctrl := gomock.NewController(t)
	tm := transaction.NewMockTransactionManager(ctrl)
	uow := transaction.NewMockUnitOfWork(ctrl)
	outboxRepository := applicationRepository.NewMockOutboxRepository(ctrl)

	uow.EXPECT().OutboxRepository().Return(outboxRepository)
	outboxRepository.EXPECT().Enqueue(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, message *model.OutboxMessage) (*model.OutboxMessage, error) {
		if !errors.Is(ctx.Err(), context.Canceled) {
			t.Fatalf("expected canceled context, got %v", ctx.Err())
		}
		return message, nil
	})

	tm.EXPECT().WithinTx(gomock.Any(), transaction.EmptyOpts(), gomock.Any()).DoAndReturn(func(ctx context.Context, _ transaction.TransactionOpts, fn transaction.TransactionCallback) error {
		if !errors.Is(ctx.Err(), context.Canceled) {
			t.Fatalf("expected canceled context in transaction manager, got %v", ctx.Err())
		}
		return fn(ctx, uow)
	})

	subject := NewHelloProducerWorker(tm, config.HelloProducerConfig{Topic: "topic"})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := subject.produceMessage(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
