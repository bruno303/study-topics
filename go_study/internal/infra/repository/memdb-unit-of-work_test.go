package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/bruno303/study-topics/go-study/internal/application/hello/models"
	"github.com/bruno303/study-topics/go-study/internal/application/transaction"
)

func TestMemDbUnitOfWork_WithinTx_UsesProvidedContextAndAccessor(t *testing.T) {
	uow := NewMemDbUnitOfWork(nil)
	ctx := context.WithValue(t.Context(), "scope", "memdb")
	callbackCalled := false

	err := uow.WithinTx(ctx, func(gotCtx context.Context, repos transaction.RepositoryAccessor) error {
		callbackCalled = true
		if gotCtx != ctx {
			t.Fatalf("expected callback to receive original context")
		}
		if repos != uow {
			t.Fatalf("expected callback to receive unit of work as repository accessor")
		}
		if repos.HelloRepository() == nil {
			t.Fatal("expected hello repository to be initialized")
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

func TestMemDbUnitOfWork_WithinTx_WhenCallbackReturnsError_PropagatesError(t *testing.T) {
	uow := NewMemDbUnitOfWork(nil)
	expectedErr := errors.New("callback failed")

	err := uow.WithinTx(t.Context(), func(context.Context, transaction.RepositoryAccessor) error {
		return expectedErr
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected callback error %v, got %v", expectedErr, err)
	}
}

func TestNewMemDbUnitOfWork_WhenStorageIsShared_SharesStateAcrossTransactions(t *testing.T) {
	storage := NewMemDbStorage()
	uow1 := NewMemDbUnitOfWork(storage)
	uow2 := NewMemDbUnitOfWork(storage)
	entity := models.HelloData{Id: "shared-id", Name: "Shared", Age: 10}

	err := uow1.WithinTx(context.Background(), func(ctx context.Context, repos transaction.RepositoryAccessor) error {
		_, err := repos.HelloRepository().Save(ctx, &entity)
		return err
	})
	if err != nil {
		t.Fatalf("expected save without error, got %v", err)
	}

	var list []models.HelloData
	err = uow2.WithinTx(context.Background(), func(ctx context.Context, repos transaction.RepositoryAccessor) error {
		var err error
		list, err = repos.HelloRepository().ListAll(ctx)
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
