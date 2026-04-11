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

	if _, ok := repoContainer.UnitOfWork.(*repositoryinfra.MemDbUnitOfWork); !ok {
		t.Fatalf("expected memdb unit of work, got %T", repoContainer.UnitOfWork)
	}

	err := repoContainer.UnitOfWork.WithinTx(context.Background(), func(ctx context.Context, repos transaction.RepositoryAccessor) error {
		_, err := repos.HelloRepository().Save(ctx, &entity)
		return err
	})
	if err != nil {
		t.Fatalf("expected memdb save transaction without error, got %v", err)
	}

	var list []models.HelloData
	err = repoContainer.UnitOfWork.WithinTx(context.Background(), func(ctx context.Context, repos transaction.RepositoryAccessor) error {
		var err error
		list, err = repos.HelloRepository().ListAll(ctx)
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

func TestNewRepositoryContainer_WhenDriverIsPgxPool_UsesPgxUnitOfWork(t *testing.T) {
	cfg := &config.Config{}
	cfg.Database.Driver = config.DatabaseDriverPGXPool

	repoContainer := newRepositoryContainer(cfg, nil)

	if _, ok := repoContainer.UnitOfWork.(*repositoryinfra.PgxUnitOfWork); !ok {
		t.Fatalf("expected pgx unit of work, got %T", repoContainer.UnitOfWork)
	}
}
