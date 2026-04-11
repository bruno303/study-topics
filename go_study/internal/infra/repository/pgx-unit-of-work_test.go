package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	applicationModels "github.com/bruno303/study-topics/go-study/internal/application/hello/models"
	applicationRepository "github.com/bruno303/study-topics/go-study/internal/application/repository"
	"github.com/bruno303/study-topics/go-study/internal/application/transaction"
	"github.com/bruno303/study-topics/go-study/tests/integration"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type fakeTx struct {
	commitFn      func(context.Context) error
	rollbackFn    func(context.Context) error
	commitCalls   int
	rollbackCalls int
}

type fakeRepositoryAccessor struct {
	helloRepository applicationRepository.HelloRepository
}

type fakeHelloRepository struct{}

func (f *fakeTx) Begin(context.Context) (pgx.Tx, error) { return nil, errors.New("not implemented") }
func (f *fakeTx) Commit(ctx context.Context) error {
	f.commitCalls++
	if f.commitFn != nil {
		return f.commitFn(ctx)
	}
	return nil
}
func (f *fakeTx) Rollback(ctx context.Context) error {
	f.rollbackCalls++
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

func (f fakeRepositoryAccessor) HelloRepository() applicationRepository.HelloRepository {
	return f.helloRepository
}

func (fakeHelloRepository) Save(context.Context, *applicationModels.HelloData) (*applicationModels.HelloData, error) {
	return nil, nil
}

func (fakeHelloRepository) ListAll(context.Context) ([]applicationModels.HelloData, error) {
	return nil, nil
}

func TestPgxUnitOfWork_WithinTx_WhenBeginFails_DoesNotInvokeCallback(t *testing.T) {
	pool := integration.SetupTestDB(t)
	defer integration.CleanupTestDB(t, pool)
	pool.Close()

	uow := NewPgxUnitOfWork(&PgxUnitOfWorkConfig{Pool: pool})
	callbackCalled := false

	err := uow.WithinTx(t.Context(), func(context.Context, transaction.RepositoryAccessor) error {
		callbackCalled = true
		return nil
	})

	if err == nil {
		t.Fatal("expected begin error, got nil")
	}
	if callbackCalled {
		t.Fatal("expected callback not to be called")
	}
}

func TestPgxUnitOfWork_WithinTx_WhenBeginSucceeds_PassesContextAndRepositoryAccessor(t *testing.T) {
	pool := integration.SetupTestDB(t)
	defer integration.CleanupTestDB(t, pool)

	uow := NewPgxUnitOfWork(&PgxUnitOfWorkConfig{Pool: pool})
	ctx := context.WithValue(t.Context(), struct{}{}, "pgx")
	entityID := fmt.Sprintf("uow-commit-%d", time.Now().UnixNano())
	callbackCalled := false

	err := uow.WithinTx(ctx, func(gotCtx context.Context, repos transaction.RepositoryAccessor) error {
		callbackCalled = true
		if gotCtx != ctx {
			t.Fatalf("expected callback to receive transaction context")
		}
		if repos == nil {
			t.Fatal("expected repository accessor, got nil")
		}
		helloRepository := repos.HelloRepository()
		if helloRepository == nil {
			t.Fatal("expected hello repository to be initialized")
		}

		_, err := helloRepository.Save(gotCtx, &applicationModels.HelloData{Id: entityID, Name: "Committed", Age: 31})
		return err
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !callbackCalled {
		t.Fatal("expected callback to be called")
	}
	if count := countHelloRowsByID(t, pool, entityID); count != 1 {
		t.Fatalf("expected committed row to exist, got count %d", count)
	}
}

func TestPgxUnitOfWork_WithinTx_WhenCallbackReturnsError_RollsBackAndReturnsCallbackError(t *testing.T) {
	pool := integration.SetupTestDB(t)
	defer integration.CleanupTestDB(t, pool)

	uow := NewPgxUnitOfWork(&PgxUnitOfWorkConfig{Pool: pool})
	entityID := fmt.Sprintf("uow-rollback-%d", time.Now().UnixNano())
	callbackErr := errors.New("callback failed")

	err := uow.WithinTx(t.Context(), func(ctx context.Context, repos transaction.RepositoryAccessor) error {
		_, err := repos.HelloRepository().Save(ctx, &applicationModels.HelloData{Id: entityID, Name: "RolledBack", Age: 22})
		if err != nil {
			return err
		}
		return callbackErr
	})

	if !errors.Is(err, callbackErr) {
		t.Fatalf("expected callback error %v, got %v", callbackErr, err)
	}
	if count := countHelloRowsByID(t, pool, entityID); count != 0 {
		t.Fatalf("expected rolled back row to be absent, got count %d", count)
	}
}

func TestPgxUnitOfWork_Execute_WhenCallbackAndRollbackFail_ReturnsJoinedError(t *testing.T) {
	callbackErr := errors.New("callback failed")
	rollbackErr := errors.New("rollback failed")
	tx := &fakeTx{rollbackFn: func(context.Context) error { return rollbackErr }}
	uow := NewPgxUnitOfWork(&PgxUnitOfWorkConfig{})

	err := uow.execute(t.Context(), tx, fakeRepositoryAccessor{helloRepository: fakeHelloRepository{}}, func(context.Context, transaction.RepositoryAccessor) error {
		return callbackErr
	})

	if !errors.Is(err, callbackErr) {
		t.Fatalf("expected joined error to contain callback error, got %v", err)
	}
	if !errors.Is(err, rollbackErr) {
		t.Fatalf("expected joined error to contain rollback error, got %v", err)
	}
	if !strings.Contains(err.Error(), "rollback transaction") {
		t.Fatalf("expected rollback context in error, got %v", err)
	}
}

func TestPgxUnitOfWork_Execute_WhenCommitFails_ReturnsCommitError(t *testing.T) {
	commitErr := errors.New("commit failed")
	tx := &fakeTx{commitFn: func(context.Context) error { return commitErr }}
	uow := NewPgxUnitOfWork(&PgxUnitOfWorkConfig{})

	err := uow.execute(t.Context(), tx, fakeRepositoryAccessor{helloRepository: fakeHelloRepository{}}, func(context.Context, transaction.RepositoryAccessor) error {
		return nil
	})

	if !errors.Is(err, commitErr) {
		t.Fatalf("expected commit error %v, got %v", commitErr, err)
	}
	if tx.rollbackCalls != 0 {
		t.Fatalf("expected no rollback calls, got %d", tx.rollbackCalls)
	}
}

func TestPgxUnitOfWork_Execute_WhenCallbackPanics_RollsBackAndRePanicsOriginalValue(t *testing.T) {
	tx := &fakeTx{}
	uow := NewPgxUnitOfWork(&PgxUnitOfWorkConfig{})
	expectedPanic := "boom"

	var recovered any
	func() {
		defer func() {
			recovered = recover()
		}()

		_ = uow.execute(t.Context(), tx, fakeRepositoryAccessor{helloRepository: fakeHelloRepository{}}, func(context.Context, transaction.RepositoryAccessor) error {
			panic(expectedPanic)
		})
	}()

	if recovered != expectedPanic {
		t.Fatalf("expected panic %q, got %#v", expectedPanic, recovered)
	}
	if tx.rollbackCalls != 1 {
		t.Fatalf("expected one rollback call, got %d", tx.rollbackCalls)
	}
	if tx.commitCalls != 0 {
		t.Fatalf("expected no commit calls, got %d", tx.commitCalls)
	}
}

func TestCombineCallbackAndRollbackErr_WhenRollbackNil_ReturnsCallbackError(t *testing.T) {
	callbackErr := errors.New("callback failed")

	got := combineCallbackAndRollbackErr(callbackErr, nil)
	if !errors.Is(got, callbackErr) {
		t.Fatalf("expected callback error to be preserved, got %v", got)
	}
}

func countHelloRowsByID(t *testing.T, pool interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}, id string) int {
	t.Helper()

	var count int
	err := pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM HELLO_DATA WHERE ID = $1", id).Scan(&count)
	if err != nil {
		t.Fatalf("failed to count hello rows: %v", err)
	}

	return count
}
