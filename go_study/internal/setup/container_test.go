package setup

import (
	"context"
	"testing"

	"github.com/bruno303/study-topics/go-study/internal/application/hello/models"
	"github.com/bruno303/study-topics/go-study/internal/application/transaction"
	"github.com/bruno303/study-topics/go-study/internal/config"
	repositoryinfra "github.com/bruno303/study-topics/go-study/internal/infra/repository"
)

func TestNewRepositoryContainer_WhenDriverIsMemDb_UsesSharedMemDbPathAcrossTransactions(t *testing.T) {
	cfg := &config.Config{}
	cfg.Database.Driver = config.DatabaseDriverMemDB

	repoContainer := newRepositoryContainer(cfg, nil)
	entity := models.HelloData{Id: "container-memdb-shared", Name: "MemDb", Age: 20}

	if _, ok := repoContainer.TransactionManager.(*repositoryinfra.MemDbTransactionManager); !ok {
		t.Fatalf("expected memdb transaction manager, got %T", repoContainer.TransactionManager)
	}

	err := repoContainer.TransactionManager.WithinTx(context.Background(), transaction.EmptyOpts(), func(ctx context.Context, uow transaction.UnitOfWork) error {
		_, err := uow.HelloRepository().Save(ctx, &entity)
		return err
	})
	if err != nil {
		t.Fatalf("expected memdb save transaction without error, got %v", err)
	}

	var list []models.HelloData
	err = repoContainer.TransactionManager.WithinTx(context.Background(), transaction.EmptyOpts(), func(ctx context.Context, uow transaction.UnitOfWork) error {
		var err error
		list, err = uow.HelloRepository().ListAll(ctx)
		return err
	})
	if err != nil {
		t.Fatalf("expected memdb list transaction without error, got %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected one entity in list, got %d", len(list))
	}
	if list[0] != entity {
		t.Fatalf("expected listed entity %+v, got %+v", entity, list[0])
	}
}

func TestNewRepositoryContainer_WhenDriverIsPgxPool_UsesPgxTransactionManager(t *testing.T) {
	cfg := &config.Config{}
	cfg.Database.Driver = config.DatabaseDriverPGXPool

	repoContainer := newRepositoryContainer(cfg, nil)

	if _, ok := repoContainer.TransactionManager.(*repositoryinfra.PgxTransactionManager); !ok {
		t.Fatalf("expected pgx transaction manager, got %T", repoContainer.TransactionManager)
	}
}
