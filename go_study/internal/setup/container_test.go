package setup

import (
	"context"
	"testing"

	"github.com/bruno303/study-topics/go-study/internal/application/hello/models"
	"github.com/bruno303/study-topics/go-study/internal/application/transaction"
	"github.com/bruno303/study-topics/go-study/internal/config"
)

func TestNewRepositoryContainer_WhenDriverIsMemDb_UsesSharedMemDbPathAcrossTransactions(t *testing.T) {
	cfg := &config.Config{}
	cfg.Database.Driver = config.DatabaseDriverMemDB

	repoContainer := newRepositoryContainer(cfg, nil)

	entity := &models.HelloData{Id: "container-memdb-shared", Name: "MemDb", Age: 20}

	err := repoContainer.TransactionManager.WithinTx(context.Background(), func(ctx context.Context, uow transaction.UnitOfWork) error {
		repo := uow.HelloRepository()
		if repo == nil {
			t.Fatal("expected hello repository to be available")
		}

		if _, saveErr := repo.Save(ctx, entity); saveErr != nil {
			t.Fatalf("expected save without error, got %v", saveErr)
		}

		return nil
	})
	if err != nil {
		t.Fatalf("expected memdb save transaction without error, got %v", err)
	}

	err = repoContainer.TransactionManager.WithinTx(context.Background(), func(ctx context.Context, uow transaction.UnitOfWork) error {
		repo := uow.HelloRepository()
		if repo == nil {
			t.Fatal("expected hello repository to be available")
		}

		list, listErr := repo.ListAll(ctx)
		if listErr != nil {
			t.Fatalf("expected list without error, got %v", listErr)
		}

		if len(list) != 1 {
			t.Fatalf("expected one entity in list, got %d", len(list))
		}

		if list[0] != *entity {
			t.Fatalf("expected listed entity %+v, got %+v", *entity, list[0])
		}

		return nil
	})
	if err != nil {
		t.Fatalf("expected memdb list transaction without error, got %v", err)
	}
}

func TestNewRepositoryContainer_WhenDriverIsPgxPool_UsesPgxPath(t *testing.T) {
	cfg := &config.Config{}
	cfg.Database.Driver = config.DatabaseDriverPGXPool

	repoContainer := newRepositoryContainer(cfg, nil)

	callbackCalled := false
	var gotErr error
	var recovered any

	func() {
		defer func() {
			recovered = recover()
		}()

		gotErr = repoContainer.TransactionManager.WithinTx(context.Background(), func(_ context.Context, _ transaction.UnitOfWork) error {
			callbackCalled = true
			return nil
		})
	}()

	if callbackCalled {
		t.Fatal("expected callback not to be called when pgx begin cannot start")
	}

	if recovered == nil && gotErr == nil {
		t.Fatal("expected pgx path to fail with nil pool (error or panic), got success")
	}
}
