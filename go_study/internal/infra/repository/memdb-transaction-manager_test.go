package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/bruno303/study-topics/go-study/internal/application/hello/models"
	"github.com/bruno303/study-topics/go-study/internal/application/transaction"
	"go.uber.org/mock/gomock"
)

func TestMemDbTransactionManager_WithinTx_UsesProvidedContextAndUnitOfWork(t *testing.T) {
	tm := NewMemDbTransactionManager(nil)
	ctx := context.WithValue(t.Context(), "scope", "memdb")
	callbackCalled := false

	err := tm.WithinTx(ctx, transaction.EmptyOpts(), func(gotCtx context.Context, uow transaction.UnitOfWork) error {
		callbackCalled = true
		if gotCtx != ctx {
			t.Fatalf("expected callback to receive original context")
		}
		if uow == nil {
			t.Fatal("expected callback to receive unit of work")
		}
		if uow.HelloRepository() == nil {
			t.Fatal("expected hello repository to be initialized")
		}
		if uow.OutboxRepository() == nil {
			t.Fatal("expected outbox repository to be initialized")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !callbackCalled {
		t.Fatal("expected callback to be called")
	}
}

func TestMemDbTransactionManager_WithinTx_WhenCallbackReturnsError_PropagatesError(t *testing.T) {
	tm := NewMemDbTransactionManager(nil)
	expectedErr := errors.New("callback failed")

	err := tm.WithinTx(t.Context(), transaction.EmptyOpts(), func(context.Context, transaction.UnitOfWork) error {
		return expectedErr
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected callback error %v, got %v", expectedErr, err)
	}
}

func TestMemDbTransactionManager_WithinTx_WhenParentProvided_ReusesParentUnitOfWork(t *testing.T) {
	ctrl := gomock.NewController(t)
	parent := transaction.NewMockUnitOfWork(ctrl)
	tm := NewMemDbTransactionManager(nil)
	ctx := context.WithValue(t.Context(), "scope", "parent")
	callbackCalled := false

	err := tm.WithinTx(ctx, transaction.TransactionOpts{Parent: parent}, func(gotCtx context.Context, uow transaction.UnitOfWork) error {
		callbackCalled = true
		if gotCtx != ctx {
			t.Fatalf("expected callback to receive original context")
		}
		if uow != parent {
			t.Fatalf("expected parent unit of work, got %T", uow)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !callbackCalled {
		t.Fatal("expected callback to be called")
	}
}

func TestNewMemDbTransactionManager_WhenStorageIsShared_SharesStateAcrossTransactions(t *testing.T) {
	storage := NewMemDbStorage()
	tm1 := NewMemDbTransactionManager(storage)
	tm2 := NewMemDbTransactionManager(storage)
	entity := models.HelloData{Id: "shared-id", Name: "Shared", Age: 10}

	err := tm1.WithinTx(context.Background(), transaction.EmptyOpts(), func(ctx context.Context, uow transaction.UnitOfWork) error {
		_, err := uow.HelloRepository().Save(ctx, &entity)
		return err
	})
	if err != nil {
		t.Fatalf("expected save without error, got %v", err)
	}

	var list []models.HelloData
	err = tm2.WithinTx(context.Background(), transaction.EmptyOpts(), func(ctx context.Context, uow transaction.UnitOfWork) error {
		var err error
		list, err = uow.HelloRepository().ListAll(ctx)
		return err
	})
	if err != nil {
		t.Fatalf("expected list without error, got %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected one entity in list, got %d", len(list))
	}
	if list[0] != entity {
		t.Fatalf("expected listed entity %+v, got %+v", entity, list[0])
	}
}
