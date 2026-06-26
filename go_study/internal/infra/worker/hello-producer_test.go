package worker

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/bruno303/study-topics/go-study/internal/application/outbox"
	"github.com/bruno303/study-topics/go-study/internal/application/transaction"
	"github.com/bruno303/study-topics/go-study/internal/config"

	"go.uber.org/mock/gomock"
)

func TestProduce(t *testing.T) {
	ctrl := gomock.NewController(t)
	outboxRepo := outbox.NewMockOutboxRepository(ctrl)
	uow := transaction.NewMockUnitOfWork(ctrl)
	txManager := NewMockTransactionManager(ctrl)

	uow.EXPECT().OutboxRepository().Return(outboxRepo).Times(1)
	outboxRepo.EXPECT().Insert(gomock.Any(), gomock.Any()).Return(nil).Times(1)

	txManager.
		EXPECT().
		WithinTx(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, opts transaction.TransactionOpts, fn transaction.TransactionCallback) error {
			return fn(ctx, uow)
		}).Times(1)

	subject := NewHelloProducerWorker(txManager, config.HelloProducerConfig{
		IntervalMillis: time.Hour.Milliseconds(),
		Topic:          "topic",
		Enabled:        true,
	})

	if err := subject.produceMessage(context.Background()); err != nil {
		t.Errorf("Error '%v' was not expected", err)
	}
}

func TestProduceWithError(t *testing.T) {
	ctrl := gomock.NewController(t)
	outboxRepo := outbox.NewMockOutboxRepository(ctrl)
	uow := transaction.NewMockUnitOfWork(ctrl)
	txManager := NewMockTransactionManager(ctrl)

	uow.EXPECT().OutboxRepository().Return(outboxRepo).Times(1)
	outboxRepo.EXPECT().Insert(gomock.Any(), gomock.Any()).Return(errors.New("error")).Times(1)

	txManager.
		EXPECT().
		WithinTx(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, opts transaction.TransactionOpts, fn transaction.TransactionCallback) error {
			return fn(ctx, uow)
		}).Times(1)

	subject := NewHelloProducerWorker(txManager, config.HelloProducerConfig{
		IntervalMillis: time.Hour.Milliseconds(),
		Topic:          "topic",
		Enabled:        true,
	})

	err := subject.produceMessage(context.Background())

	if err == nil || err.Error() != "error" {
		t.Errorf("Error 'error' was expected, got '%v'", err)
	}
}

func TestProduceWithTxError(t *testing.T) {
	ctrl := gomock.NewController(t)
	txManager := NewMockTransactionManager(ctrl)

	txManager.
		EXPECT().
		WithinTx(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(errors.New("tx error")).
		Times(1)

	subject := NewHelloProducerWorker(txManager, config.HelloProducerConfig{
		IntervalMillis: time.Hour.Milliseconds(),
		Topic:          "topic",
		Enabled:        true,
	})

	err := subject.produceMessage(context.Background())

	if err == nil || err.Error() != "tx error" {
		t.Errorf("Error 'tx error' was expected, got '%v'", err)
	}
}
