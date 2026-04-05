package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type fakeTx struct {
	commitFn   func(context.Context) error
	rollbackFn func(context.Context) error
}

func (f *fakeTx) Begin(context.Context) (pgx.Tx, error) { return nil, errors.New("not implemented") }
func (f *fakeTx) Commit(ctx context.Context) error {
	if f.commitFn != nil {
		return f.commitFn(ctx)
	}
	return nil
}
func (f *fakeTx) Rollback(ctx context.Context) error {
	if f.rollbackFn != nil {
		return f.rollbackFn(ctx)
	}
	return nil
}
func (f *fakeTx) CopyFrom(context.Context, pgx.Identifier, []string, pgx.CopyFromSource) (int64, error) {
	return 0, errors.New("not implemented")
}
func (f *fakeTx) SendBatch(context.Context, *pgx.Batch) pgx.BatchResults { return nil }
func (f *fakeTx) LargeObjects() pgx.LargeObjects                         { return pgx.LargeObjects{} }
func (f *fakeTx) Prepare(context.Context, string, string) (*pgconn.StatementDescription, error) {
	return nil, errors.New("not implemented")
}
func (f *fakeTx) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, errors.New("not implemented")
}
func (f *fakeTx) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return nil, errors.New("not implemented")
}
func (f *fakeTx) QueryRow(context.Context, string, ...any) pgx.Row { return nil }
func (f *fakeTx) Conn() *pgx.Conn                                  { return nil }

func TestPgxUnitOfWork_Begin_WhenTransactionAlreadyOpen_ReturnsError(t *testing.T) {
	uow := newPgxUnitOfWorkWithTxRef(&PgxUnitOfWorkConfig{Pool: nil}, &transactionRef{tx: &fakeTx{}})

	err := uow.Begin(context.Background())
	if !errors.Is(err, TransactionAlreadyOpenErr) {
		t.Fatalf("expected TransactionAlreadyOpenErr, got: %v", err)
	}
}

func TestPgxUnitOfWork_HelloRepository_ReturnsRepository(t *testing.T) {
	uow := NewPgxUnitOfWork(&PgxUnitOfWorkConfig{Pool: nil})

	if repo := uow.HelloRepository(); repo == nil {
		t.Fatal("expected hello repository to be initialized")
	}
}

func TestPgxUnitOfWork_Commit_WhenTransactionNotOpened_ReturnsError(t *testing.T) {
	uow := NewPgxUnitOfWork(&PgxUnitOfWorkConfig{Pool: nil})

	err := uow.Commit(context.Background())
	if !errors.Is(err, TransactionNotOpenedErr) {
		t.Fatalf("expected TransactionNotOpenedErr, got: %v", err)
	}
}

func TestPgxUnitOfWork_Commit_WhenTransactionOpened_CommitsAndClearsCurrentTransaction(t *testing.T) {
	committed := false
	tx := &fakeTx{commitFn: func(context.Context) error {
		committed = true
		return nil
	}}
	uow := newPgxUnitOfWorkWithTxRef(&PgxUnitOfWorkConfig{Pool: nil}, &transactionRef{tx: tx})

	err := uow.Commit(context.Background())
	if err != nil {
		t.Fatalf("expected commit without error, got: %v", err)
	}
	if !committed {
		t.Fatal("expected commit to be called on current transaction")
	}
	if uow.txRef.current() != nil {
		t.Fatal("expected current transaction to be cleared after commit")
	}
}

func TestPgxUnitOfWork_Rollback_WhenTransactionNotOpened_ReturnsError(t *testing.T) {
	uow := NewPgxUnitOfWork(&PgxUnitOfWorkConfig{Pool: nil})

	err := uow.Rollback(context.Background())
	if !errors.Is(err, TransactionNotOpenedErr) {
		t.Fatalf("expected TransactionNotOpenedErr, got: %v", err)
	}
}

func TestPgxUnitOfWork_Rollback_WhenTransactionOpened_RollsBackAndClearsCurrentTransaction(t *testing.T) {
	rolledBack := false
	tx := &fakeTx{rollbackFn: func(context.Context) error {
		rolledBack = true
		return nil
	}}
	uow := newPgxUnitOfWorkWithTxRef(&PgxUnitOfWorkConfig{Pool: nil}, &transactionRef{tx: tx})

	err := uow.Rollback(context.Background())
	if err != nil {
		t.Fatalf("expected rollback without error, got: %v", err)
	}
	if !rolledBack {
		t.Fatal("expected rollback to be called on current transaction")
	}
	if uow.txRef.current() != nil {
		t.Fatal("expected current transaction to be cleared after rollback")
	}
}
