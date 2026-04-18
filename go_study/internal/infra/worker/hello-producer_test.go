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

func TestProduce(t *testing.T) {
	ctrl := gomock.NewController(t)
	transactionManager := transaction.NewMockTransactionManager(ctrl)
	uow := transaction.NewMockUnitOfWork(ctrl)
	outboxRepo := applicationRepository.NewMockOutboxRepository(ctrl)

	transactionManager.
		EXPECT().
		WithinTx(gomock.Any(), transaction.EmptyOpts(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, _ transaction.TransactionOpts, fn transaction.TransactionCallback) error {
			return fn(ctx, uow)
		}).Times(1)

	uow.EXPECT().OutboxRepository().Return(outboxRepo).Times(1)
	outboxRepo.EXPECT().Enqueue(gomock.Any(), gomock.AssignableToTypeOf(applicationRepository.OutboxMessage{})).Return(nil).Times(1)

	subject := NewHelloProducerWorker(transactionManager, config.HelloProducerConfig{
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
	transactionManager := transaction.NewMockTransactionManager(ctrl)
	uow := transaction.NewMockUnitOfWork(ctrl)
	outboxRepo := applicationRepository.NewMockOutboxRepository(ctrl)
	expectedErr := errors.New("error")

	transactionManager.
		EXPECT().
		WithinTx(gomock.Any(), transaction.EmptyOpts(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, _ transaction.TransactionOpts, fn transaction.TransactionCallback) error {
			return fn(ctx, uow)
		}).Times(1)

	uow.EXPECT().OutboxRepository().Return(outboxRepo).Times(1)
	outboxRepo.EXPECT().Enqueue(gomock.Any(), gomock.AssignableToTypeOf(applicationRepository.OutboxMessage{})).Return(expectedErr).Times(1)

	subject := NewHelloProducerWorker(transactionManager, config.HelloProducerConfig{
		IntervalMillis: time.Hour.Milliseconds(),
		Topic:          "topic",
		Enabled:        true,
	})

	err := subject.produceMessage(context.Background())

	if !errors.Is(err, expectedErr) {
		t.Errorf("Error 'error' was expected")
	}
}
