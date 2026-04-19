package setup

import (
	"context"
	"errors"
	"testing"

	"github.com/bruno303/study-topics/go-study/internal/application/hello/models"
	"github.com/bruno303/study-topics/go-study/internal/application/transaction"
	"github.com/bruno303/study-topics/go-study/internal/config"
	repositoryinfra "github.com/bruno303/study-topics/go-study/internal/infra/repository"
	"github.com/jackc/pgx/v5/pgxpool"
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

func TestNewDatabasePool_WhenDriverIsMemDB_SkipsMigrationAndConnection(t *testing.T) {
	originalRunMigrations := runDatabaseMigrations
	originalConnectDatabase := connectDatabase
	t.Cleanup(func() {
		runDatabaseMigrations = originalRunMigrations
		connectDatabase = originalConnectDatabase
	})

	runDatabaseMigrations = func(ctx context.Context, cfg *config.Config) error {
		t.Fatal("runDatabaseMigrations should not be called for memdb")
		return nil
	}
	connectDatabase = func(cfg *config.Config) *pgxpool.Pool {
		t.Fatal("connectDatabase should not be called for memdb")
		return nil
	}

	cfg := &config.Config{}
	cfg.Database.Driver = config.DatabaseDriverMemDB

	if pool := newDatabasePool(context.Background(), cfg); pool != nil {
		t.Fatalf("expected nil pool for memdb, got %v", pool)
	}
}

func TestNewDatabasePool_WhenDriverIsPgxPool_RunsMigrationBeforeConnect(t *testing.T) {
	originalRunMigrations := runDatabaseMigrations
	originalConnectDatabase := connectDatabase
	t.Cleanup(func() {
		runDatabaseMigrations = originalRunMigrations
		connectDatabase = originalConnectDatabase
	})

	callOrder := []string{}
	runDatabaseMigrations = func(ctx context.Context, cfg *config.Config) error {
		callOrder = append(callOrder, "migrations")
		return nil
	}
	connectDatabase = func(cfg *config.Config) *pgxpool.Pool {
		callOrder = append(callOrder, "connect")
		return nil
	}

	cfg := &config.Config{}
	cfg.Database.Driver = config.DatabaseDriverPGXPool

	if pool := newDatabasePool(context.Background(), cfg); pool != nil {
		t.Fatalf("expected nil stub pool, got %v", pool)
	}
	if len(callOrder) != 2 || callOrder[0] != "migrations" || callOrder[1] != "connect" {
		t.Fatalf("unexpected call order: %v", callOrder)
	}
}

func TestNewDatabasePool_WhenMigrationFailsPanics(t *testing.T) {
	originalRunMigrations := runDatabaseMigrations
	originalConnectDatabase := connectDatabase
	t.Cleanup(func() {
		runDatabaseMigrations = originalRunMigrations
		connectDatabase = originalConnectDatabase
	})

	runDatabaseMigrations = func(ctx context.Context, cfg *config.Config) error {
		return errors.New("boom")
	}
	connectDatabase = func(cfg *config.Config) *pgxpool.Pool {
		t.Fatal("connectDatabase should not be called when migrations fail")
		return nil
	}

	cfg := &config.Config{}
	cfg.Database.Driver = config.DatabaseDriverPGXPool

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic from migration failure")
		}
	}()

	_ = newDatabasePool(context.Background(), cfg)
}
