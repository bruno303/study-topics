package worker

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/bruno303/study-topics/go-study/internal/application/outbox"
	"github.com/bruno303/study-topics/go-study/internal/application/transaction"
	"github.com/bruno303/study-topics/go-study/internal/config"

	"go.uber.org/mock/gomock"
)

type MockProducer struct {
	ctrl     *gomock.Controller
	recorder *MockProducerMockRecorder
}

type MockProducerMockRecorder struct {
	mock *MockProducer
}

func NewMockProducer(ctrl *gomock.Controller) *MockProducer {
	mock := &MockProducer{ctrl: ctrl}
	mock.recorder = &MockProducerMockRecorder{mock}
	return mock
}

func (m *MockProducer) EXPECT() *MockProducerMockRecorder {
	return m.recorder
}

func (m *MockProducer) ProduceSync(ctx context.Context, msg, topic string, headers map[string]string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ProduceSync", ctx, msg, topic, headers)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockProducerMockRecorder) ProduceSync(ctx, msg, topic, headers any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ProduceSync", reflect.TypeOf((*MockProducer)(nil).ProduceSync), ctx, msg, topic, headers)
}

func TestProcessBatch_EmptyBatch(t *testing.T) {
	ctrl := gomock.NewController(t)

	outboxRepo := outbox.NewMockOutboxRepository(ctrl)
	uow := transaction.NewMockUnitOfWork(ctrl)
	txManager := NewMockTransactionManager(ctrl)
	prod := NewMockProducer(ctrl)

	uow.EXPECT().OutboxRepository().Return(outboxRepo).Times(1)
	outboxRepo.EXPECT().FetchPendingBatch(gomock.Any(), 100).Return([]*outbox.OutboxMessage{}, nil).Times(1)

	txManager.
		EXPECT().
		WithinTx(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, opts transaction.TransactionOpts, fn transaction.TransactionCallback) error {
			return fn(ctx, uow)
		}).Times(1)

	subject := NewOutboxRelayWorker(txManager, prod, config.OutboxRelayConfig{
		Enabled:        true,
		PollIntervalMs: 5000,
		BatchSize:      100,
		MaxAttempts:    3,
	})

	if err := subject.processBatch(context.Background()); err != nil {
		t.Errorf("Error '%v' was not expected", err)
	}
}

func TestProcessBatch_Success(t *testing.T) {
	ctrl := gomock.NewController(t)

	outboxRepo := outbox.NewMockOutboxRepository(ctrl)
	uow := transaction.NewMockUnitOfWork(ctrl)
	txManager := NewMockTransactionManager(ctrl)
	prod := NewMockProducer(ctrl)

	msg := &outbox.OutboxMessage{
		ID:      "msg-1",
		Payload: `{"key":"value"}`,
		Topic:   "test-topic",
		Headers: `{"h1":"v1"}`,
	}

	uow.EXPECT().OutboxRepository().Return(outboxRepo).Times(1)
	outboxRepo.EXPECT().FetchPendingBatch(gomock.Any(), 100).Return([]*outbox.OutboxMessage{msg}, nil).Times(1)
	prod.EXPECT().ProduceSync(gomock.Any(), `{"key":"value"}`, "test-topic", map[string]string{"h1": "v1"}).Return(nil).Times(1)
	outboxRepo.EXPECT().MarkSent(gomock.Any(), "msg-1").Return(nil).Times(1)

	txManager.
		EXPECT().
		WithinTx(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, opts transaction.TransactionOpts, fn transaction.TransactionCallback) error {
			return fn(ctx, uow)
		}).Times(1)

	subject := NewOutboxRelayWorker(txManager, prod, config.OutboxRelayConfig{
		Enabled:        true,
		PollIntervalMs: 5000,
		BatchSize:      100,
		MaxAttempts:    3,
	})

	if err := subject.processBatch(context.Background()); err != nil {
		t.Errorf("Error '%v' was not expected", err)
	}
}

func TestProcessBatch_ProduceError(t *testing.T) {
	ctrl := gomock.NewController(t)

	outboxRepo := outbox.NewMockOutboxRepository(ctrl)
	uow := transaction.NewMockUnitOfWork(ctrl)
	txManager := NewMockTransactionManager(ctrl)
	prod := NewMockProducer(ctrl)

	msg := &outbox.OutboxMessage{
		ID:      "msg-1",
		Payload: `{"key":"value"}`,
		Topic:   "test-topic",
		Headers: `{}`,
	}

	uow.EXPECT().OutboxRepository().Return(outboxRepo).Times(1)
	outboxRepo.EXPECT().FetchPendingBatch(gomock.Any(), 100).Return([]*outbox.OutboxMessage{msg}, nil).Times(1)
	prod.EXPECT().ProduceSync(gomock.Any(), `{"key":"value"}`, "test-topic", map[string]string(nil)).Return(errors.New("kafka error")).Times(1)
	outboxRepo.EXPECT().MarkFailed(gomock.Any(), "msg-1", "kafka error", 3).Return(nil).Times(1)

	txManager.
		EXPECT().
		WithinTx(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, opts transaction.TransactionOpts, fn transaction.TransactionCallback) error {
			return fn(ctx, uow)
		}).Times(1)

	subject := NewOutboxRelayWorker(txManager, prod, config.OutboxRelayConfig{
		Enabled:        true,
		PollIntervalMs: 5000,
		BatchSize:      100,
		MaxAttempts:    3,
	})

	if err := subject.processBatch(context.Background()); err != nil {
		t.Errorf("Error '%v' was not expected", err)
	}
}

func TestProcessBatch_InvalidHeaders(t *testing.T) {
	ctrl := gomock.NewController(t)

	outboxRepo := outbox.NewMockOutboxRepository(ctrl)
	uow := transaction.NewMockUnitOfWork(ctrl)
	txManager := NewMockTransactionManager(ctrl)
	prod := NewMockProducer(ctrl)

	msg := &outbox.OutboxMessage{
		ID:      "msg-1",
		Payload: `{"key":"value"}`,
		Topic:   "test-topic",
		Headers: `not-json`,
	}

	uow.EXPECT().OutboxRepository().Return(outboxRepo).Times(1)
	outboxRepo.EXPECT().FetchPendingBatch(gomock.Any(), 100).Return([]*outbox.OutboxMessage{msg}, nil).Times(1)
	outboxRepo.EXPECT().MarkFailed(gomock.Any(), "msg-1", gomock.Any(), 3).Return(nil).Times(1)

	txManager.
		EXPECT().
		WithinTx(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, opts transaction.TransactionOpts, fn transaction.TransactionCallback) error {
			return fn(ctx, uow)
		}).Times(1)

	subject := NewOutboxRelayWorker(txManager, prod, config.OutboxRelayConfig{
		Enabled:        true,
		PollIntervalMs: 5000,
		BatchSize:      100,
		MaxAttempts:    3,
	})

	if err := subject.processBatch(context.Background()); err != nil {
		t.Errorf("Error '%v' was not expected", err)
	}
}

func TestProcessBatch_FetchError(t *testing.T) {
	ctrl := gomock.NewController(t)

	outboxRepo := outbox.NewMockOutboxRepository(ctrl)
	uow := transaction.NewMockUnitOfWork(ctrl)
	txManager := NewMockTransactionManager(ctrl)
	prod := NewMockProducer(ctrl)

	uow.EXPECT().OutboxRepository().Return(outboxRepo).Times(1)
	outboxRepo.EXPECT().FetchPendingBatch(gomock.Any(), 100).Return(nil, errors.New("db error")).Times(1)

	txManager.
		EXPECT().
		WithinTx(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, opts transaction.TransactionOpts, fn transaction.TransactionCallback) error {
			return fn(ctx, uow)
		}).Times(1)

	subject := NewOutboxRelayWorker(txManager, prod, config.OutboxRelayConfig{
		Enabled:        true,
		PollIntervalMs: 5000,
		BatchSize:      100,
		MaxAttempts:    3,
	})

	if err := subject.processBatch(context.Background()); err == nil {
		t.Errorf("Error was expected")
	}
}

func TestProcessBatch_TxError(t *testing.T) {
	ctrl := gomock.NewController(t)
	txManager := NewMockTransactionManager(ctrl)
	prod := NewMockProducer(ctrl)

	txManager.
		EXPECT().
		WithinTx(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(errors.New("tx error")).
		Times(1)

	subject := NewOutboxRelayWorker(txManager, prod, config.OutboxRelayConfig{
		Enabled:        true,
		PollIntervalMs: 5000,
		BatchSize:      100,
		MaxAttempts:    3,
	})

	if err := subject.processBatch(context.Background()); err == nil {
		t.Errorf("Error was expected")
	}
}
