package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/bruno303/study-topics/go-study/internal/application/hello/models"
	"github.com/bruno303/study-topics/go-study/internal/application/transaction"
	"github.com/bruno303/study-topics/go-study/tests/integration"
)

func TestPgxUnitOfWork_BeginCommit_CommitsChanges(t *testing.T) {
	pool := integration.SetupTestDB(t)
	defer integration.CleanupTestDB(t, pool)

	uow := NewPgxUnitOfWork(&PgxUnitOfWorkConfig{Pool: pool})
	ctx := context.Background()
	testID := fmt.Sprintf("uow-commit-%d", time.Now().UnixNano())

	if err := uow.Begin(ctx, transaction.EmptyOpts); err != nil {
		t.Fatalf("expected begin without error, got: %v", err)
	}

	tx := uow.txRef.current()
	if tx == nil {
		t.Fatal("expected current transaction to be set")
	}

	if _, err := tx.Exec(ctx, "INSERT INTO HELLO_DATA (ID, NAME, AGE) VALUES ($1, $2, $3)", testID, "John", 30); err != nil {
		t.Fatalf("failed to insert in tx: %v", err)
	}

	if err := uow.Commit(ctx); err != nil {
		t.Fatalf("expected commit without error, got: %v", err)
	}

	var count int
	if err := pool.QueryRow(ctx, "SELECT COUNT(*) FROM HELLO_DATA WHERE ID = $1", testID).Scan(&count); err != nil {
		t.Fatalf("failed to verify committed row: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected count 1, got %d", count)
	}
}

func TestPgxUnitOfWork_BeginRollback_RollsBackChanges(t *testing.T) {
	pool := integration.SetupTestDB(t)
	defer integration.CleanupTestDB(t, pool)

	uow := NewPgxUnitOfWork(&PgxUnitOfWorkConfig{Pool: pool})
	ctx := context.Background()
	testID := fmt.Sprintf("uow-rollback-%d", time.Now().UnixNano())

	if err := uow.Begin(ctx, transaction.EmptyOpts); err != nil {
		t.Fatalf("expected begin without error, got: %v", err)
	}

	tx := uow.txRef.current()
	if _, err := tx.Exec(ctx, "INSERT INTO HELLO_DATA (ID, NAME, AGE) VALUES ($1, $2, $3)", testID, "Jane", 25); err != nil {
		t.Fatalf("failed to insert in tx: %v", err)
	}

	if err := uow.Rollback(ctx); err != nil {
		t.Fatalf("expected rollback without error, got: %v", err)
	}

	var count int
	if err := pool.QueryRow(ctx, "SELECT COUNT(*) FROM HELLO_DATA WHERE ID = $1", testID).Scan(&count); err != nil {
		t.Fatalf("failed to verify rolled back row: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected count 0, got %d", count)
	}
}

func TestPgxUnitOfWork_WithRepository_CommitAndRollback(t *testing.T) {
	pool := integration.SetupTestDB(t)
	defer integration.CleanupTestDB(t, pool)

	ctx := context.Background()
	uow := NewPgxUnitOfWork(&PgxUnitOfWorkConfig{Pool: pool})
	repo := uow.HelloRepository()
	verifyRepo := NewHelloPgxRepository(pool)

	commitID := fmt.Sprintf("uow-repo-commit-%d", time.Now().UnixNano())
	if err := uow.Begin(ctx, transaction.EmptyOpts); err != nil {
		t.Fatalf("begin failed: %v", err)
	}
	if _, err := repo.Save(ctx, &models.HelloData{Id: commitID, Name: "Alice", Age: 33}); err != nil {
		t.Fatalf("save failed: %v", err)
	}
	if err := uow.Commit(ctx); err != nil {
		t.Fatalf("commit failed: %v", err)
	}

	if _, err := verifyRepo.FindById(ctx, commitID); err != nil {
		t.Fatalf("expected committed row to exist, got: %v", err)
	}

	rollbackID := fmt.Sprintf("uow-repo-rollback-%d", time.Now().UnixNano())
	if err := uow.Begin(ctx, transaction.EmptyOpts); err != nil {
		t.Fatalf("begin failed: %v", err)
	}
	if _, err := repo.Save(ctx, &models.HelloData{Id: rollbackID, Name: "Bob", Age: 29}); err != nil {
		t.Fatalf("save failed: %v", err)
	}
	if err := uow.Rollback(ctx); err != nil {
		t.Fatalf("rollback failed: %v", err)
	}

	if _, err := verifyRepo.FindById(ctx, rollbackID); err == nil {
		t.Fatal("expected rolled back row to be absent")
	}
}

func TestPgxUnitOfWork_Begin_ValidationErrors(t *testing.T) {
	pool := integration.SetupTestDB(t)
	defer integration.CleanupTestDB(t, pool)

	ctx := context.Background()
	uow := NewPgxUnitOfWork(&PgxUnitOfWorkConfig{Pool: pool})

	if err := uow.Begin(ctx, transaction.Opts{Propagation: transaction.Propagation(99)}); err != InvalidPropagationErr {
		t.Fatalf("expected InvalidPropagationErr, got: %v", err)
	}
}

func TestPgxUnitOfWork_CommitRollback_WithoutBegin(t *testing.T) {
	pool := integration.SetupTestDB(t)
	defer integration.CleanupTestDB(t, pool)

	ctx := context.Background()
	uow := NewPgxUnitOfWork(&PgxUnitOfWorkConfig{Pool: pool})

	if err := uow.Commit(ctx); err != TransactionNotOpenedErr {
		t.Fatalf("expected TransactionNotOpenedErr on commit, got: %v", err)
	}
	if err := uow.Rollback(ctx); err != TransactionNotOpenedErr {
		t.Fatalf("expected TransactionNotOpenedErr on rollback, got: %v", err)
	}
}

func TestPgxUnitOfWork_Begin_AlreadyOpen(t *testing.T) {
	pool := integration.SetupTestDB(t)
	defer integration.CleanupTestDB(t, pool)

	ctx := context.Background()
	uow := NewPgxUnitOfWork(&PgxUnitOfWorkConfig{Pool: pool})

	if err := uow.Begin(ctx, transaction.EmptyOpts); err != nil {
		t.Fatalf("begin failed: %v", err)
	}
	if err := uow.Begin(ctx, transaction.EmptyOpts); err != TransactionAlreadyOpenErr {
		t.Fatalf("expected TransactionAlreadyOpenErr, got: %v", err)
	}
	if err := uow.Rollback(ctx); err != nil {
		t.Fatalf("cleanup rollback failed: %v", err)
	}
}
