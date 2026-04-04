package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/bruno303/study-topics/go-study/internal/application/hello/models"
	"github.com/bruno303/study-topics/go-study/internal/application/transaction"
	"github.com/bruno303/study-topics/go-study/tests/integration"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestUnitOfWorkFactory_WithinTx_Join_UsesSharedTransactionAndRejectsFinalize(t *testing.T) {
	pool := integration.SetupTestDB(t)
	defer integration.CleanupTestDB(t, pool)

	ctx := context.Background()
	factory := NewUnitOfWorkFactory(&PgxUnitOfWorkConfig{Pool: pool})

	outerID := fmt.Sprintf("join-outer-%d", time.Now().UnixNano())
	innerID := fmt.Sprintf("join-inner-%d", time.Now().UnixNano())

	err := factory.WithinTx(ctx, transaction.RequiresNew(), func(txCtx context.Context, outer transaction.UnitOfWork) error {
		if _, saveErr := outer.HelloRepository().Save(txCtx, &models.HelloData{Id: outerID, Name: "Outer", Age: 20}); saveErr != nil {
			return saveErr
		}

		joinErr := factory.WithinTx(txCtx, transaction.Join(), func(innerCtx context.Context, joined transaction.UnitOfWork) error {
			if _, saveErr := joined.HelloRepository().Save(innerCtx, &models.HelloData{Id: innerID, Name: "Inner", Age: 21}); saveErr != nil {
				return saveErr
			}

			outerPgx, ok := outer.(*PgxUnitOfWork)
			if !ok {
				t.Fatalf("expected outer unit of work to be *PgxUnitOfWork")
			}
			if count := countHelloDataByIDTx(t, innerCtx, outerPgx.txRef.current(), innerID); count != 1 {
				t.Fatalf("expected inner insert visible inside shared tx, got count %d", count)
			}
			if count := countHelloDataByID(t, innerCtx, pool, innerID); count != 0 {
				t.Fatalf("expected uncommitted join insert to be invisible outside tx, got count %d", count)
			}

			if commitErr := joined.Commit(innerCtx); !errors.Is(commitErr, InvalidScopeTransitionErr) {
				t.Fatalf("expected InvalidScopeTransitionErr on joined commit, got: %v", commitErr)
			}
			if rollbackErr := joined.Rollback(innerCtx); !errors.Is(rollbackErr, InvalidScopeTransitionErr) {
				t.Fatalf("expected InvalidScopeTransitionErr on joined rollback, got: %v", rollbackErr)
			}

			return nil
		})
		if joinErr != nil {
			return joinErr
		}

		return nil
	})
	if err != nil {
		t.Fatalf("expected within tx join flow without error, got: %v", err)
	}

	if count := countHelloDataByID(t, ctx, pool, outerID); count != 1 {
		t.Fatalf("expected committed outer row count 1, got %d", count)
	}
	if count := countHelloDataByID(t, ctx, pool, innerID); count != 1 {
		t.Fatalf("expected committed join row count 1, got %d", count)
	}
}

func TestUnitOfWorkFactory_WithinTx_RequiresNew_IsIndependentFromOuterRollback(t *testing.T) {
	pool := integration.SetupTestDB(t)
	defer integration.CleanupTestDB(t, pool)

	ctx := context.Background()
	factory := NewUnitOfWorkFactory(&PgxUnitOfWorkConfig{Pool: pool})

	outerID := fmt.Sprintf("requires-new-outer-%d", time.Now().UnixNano())
	innerID := fmt.Sprintf("requires-new-inner-%d", time.Now().UnixNano())
	expectedOuterErr := errors.New("force outer rollback")

	err := factory.WithinTx(ctx, transaction.RequiresNew(), func(txCtx context.Context, outer transaction.UnitOfWork) error {
		if _, saveErr := outer.HelloRepository().Save(txCtx, &models.HelloData{Id: outerID, Name: "Outer", Age: 22}); saveErr != nil {
			return saveErr
		}

		innerErr := factory.WithinTx(txCtx, transaction.RequiresNew(), func(innerCtx context.Context, inner transaction.UnitOfWork) error {
			_, saveErr := inner.HelloRepository().Save(innerCtx, &models.HelloData{Id: innerID, Name: "Inner", Age: 23})
			return saveErr
		})
		if innerErr != nil {
			return innerErr
		}

		return expectedOuterErr
	})
	if !errors.Is(err, expectedOuterErr) {
		t.Fatalf("expected outer error %v, got: %v", expectedOuterErr, err)
	}

	if count := countHelloDataByID(t, ctx, pool, outerID); count != 0 {
		t.Fatalf("expected outer row rolled back, got count %d", count)
	}
	if count := countHelloDataByID(t, ctx, pool, innerID); count != 1 {
		t.Fatalf("expected requires-new inner row committed, got count %d", count)
	}
}

func TestUnitOfWorkFactory_WithinTx_Nested_SavepointCommitAndRollback(t *testing.T) {
	pool := integration.SetupTestDB(t)
	defer integration.CleanupTestDB(t, pool)

	ctx := context.Background()
	factory := NewUnitOfWorkFactory(&PgxUnitOfWorkConfig{Pool: pool})

	outerID := fmt.Sprintf("nested-outer-%d", time.Now().UnixNano())
	committedNestedID := fmt.Sprintf("nested-commit-%d", time.Now().UnixNano())
	rolledBackNestedID := fmt.Sprintf("nested-rollback-%d", time.Now().UnixNano())
	expectedNestedErr := errors.New("force nested rollback")

	err := factory.WithinTx(ctx, transaction.RequiresNew(), func(txCtx context.Context, outer transaction.UnitOfWork) error {
		if _, saveErr := outer.HelloRepository().Save(txCtx, &models.HelloData{Id: outerID, Name: "Outer", Age: 24}); saveErr != nil {
			return saveErr
		}

		if nestedErr := factory.WithinTx(txCtx, transaction.Nested(), func(nestedCtx context.Context, nested transaction.UnitOfWork) error {
			_, saveErr := nested.HelloRepository().Save(nestedCtx, &models.HelloData{Id: committedNestedID, Name: "NestedCommit", Age: 25})
			return saveErr
		}); nestedErr != nil {
			return nestedErr
		}

		nestedErr := factory.WithinTx(txCtx, transaction.Nested(), func(nestedCtx context.Context, nested transaction.UnitOfWork) error {
			if _, saveErr := nested.HelloRepository().Save(nestedCtx, &models.HelloData{Id: rolledBackNestedID, Name: "NestedRollback", Age: 26}); saveErr != nil {
				return saveErr
			}
			return expectedNestedErr
		})
		if !errors.Is(nestedErr, expectedNestedErr) {
			return fmt.Errorf("expected nested callback error %v, got: %w", expectedNestedErr, nestedErr)
		}

		outerPgx, ok := outer.(*PgxUnitOfWork)
		if !ok {
			t.Fatalf("expected outer unit of work to be *PgxUnitOfWork")
		}
		if count := countHelloDataByIDTx(t, txCtx, outerPgx.txRef.current(), committedNestedID); count != 1 {
			t.Fatalf("expected committed nested row visible in parent tx, got count %d", count)
		}
		if count := countHelloDataByIDTx(t, txCtx, outerPgx.txRef.current(), rolledBackNestedID); count != 0 {
			t.Fatalf("expected rolled back nested row to be absent in parent tx, got count %d", count)
		}

		return nil
	})
	if err != nil {
		t.Fatalf("expected nested flow to finish without error, got: %v", err)
	}

	if count := countHelloDataByID(t, ctx, pool, outerID); count != 1 {
		t.Fatalf("expected outer row committed, got count %d", count)
	}
	if count := countHelloDataByID(t, ctx, pool, committedNestedID); count != 1 {
		t.Fatalf("expected nested committed row persisted, got count %d", count)
	}
	if count := countHelloDataByID(t, ctx, pool, rolledBackNestedID); count != 0 {
		t.Fatalf("expected rolled back nested row absent, got count %d", count)
	}
}

func TestUnitOfWorkFactory_WithinTx_RequiresNew_NestedJoin_RejectsFinalizeAndPreservesLifecycle(t *testing.T) {
	pool := integration.SetupTestDB(t)
	defer integration.CleanupTestDB(t, pool)

	ctx := context.Background()
	factory := NewUnitOfWorkFactory(&PgxUnitOfWorkConfig{Pool: pool})

	nestedID := fmt.Sprintf("nested-join-nested-%d", time.Now().UnixNano())
	joinedID := fmt.Sprintf("nested-join-joined-%d", time.Now().UnixNano())

	err := factory.WithinTx(ctx, transaction.RequiresNew(), func(txCtx context.Context, outer transaction.UnitOfWork) error {
		return factory.WithinTx(txCtx, transaction.Nested(), func(nestedCtx context.Context, nested transaction.UnitOfWork) error {
			nestedPgx, ok := nested.(*PgxUnitOfWork)
			if !ok {
				t.Fatalf("expected nested unit of work to be *PgxUnitOfWork")
			}
			if !isSavepointTx(nestedPgx.txRef.current()) {
				t.Fatalf("expected nested unit of work transaction to be savepoint-backed")
			}

			if _, saveErr := nested.HelloRepository().Save(nestedCtx, &models.HelloData{Id: nestedID, Name: "Nested", Age: 30}); saveErr != nil {
				return saveErr
			}

			joinErr := factory.WithinTx(nestedCtx, transaction.Join(), func(joinCtx context.Context, joined transaction.UnitOfWork) error {
				joinedPgx, ok := joined.(*PgxUnitOfWork)
				if !ok {
					t.Fatalf("expected joined unit of work to be *PgxUnitOfWork")
				}
				if joinedPgx.txRef.current() != nestedPgx.txRef.current() {
					t.Fatalf("expected join inside nested to reuse nested transaction scope")
				}

				if _, saveErr := joined.HelloRepository().Save(joinCtx, &models.HelloData{Id: joinedID, Name: "Joined", Age: 31}); saveErr != nil {
					return saveErr
				}

				if commitErr := joined.Commit(joinCtx); !errors.Is(commitErr, InvalidScopeTransitionErr) {
					t.Fatalf("expected InvalidScopeTransitionErr on joined commit inside nested, got: %v", commitErr)
				}
				if rollbackErr := joined.Rollback(joinCtx); !errors.Is(rollbackErr, InvalidScopeTransitionErr) {
					t.Fatalf("expected InvalidScopeTransitionErr on joined rollback inside nested, got: %v", rollbackErr)
				}

				if count := countHelloDataByIDTx(t, joinCtx, nestedPgx.txRef.current(), joinedID); count != 1 {
					t.Fatalf("expected joined insert visible in nested tx scope, got count %d", count)
				}

				return nil
			})
			if joinErr != nil {
				return joinErr
			}

			if count := countHelloDataByIDTx(t, nestedCtx, nestedPgx.txRef.current(), nestedID); count != 1 {
				t.Fatalf("expected nested insert visible in nested tx scope, got count %d", count)
			}

			outerPgx, ok := outer.(*PgxUnitOfWork)
			if !ok {
				t.Fatalf("expected outer unit of work to be *PgxUnitOfWork")
			}
			if count := countHelloDataByIDTx(t, nestedCtx, outerPgx.txRef.current(), joinedID); count != 1 {
				t.Fatalf("expected joined row to remain visible in parent tx after nested join flow, got count %d", count)
			}

			return nil
		})
	})
	if err != nil {
		t.Fatalf("expected requires-new nested join flow without error, got: %v", err)
	}

	if count := countHelloDataByID(t, ctx, pool, nestedID); count != 1 {
		t.Fatalf("expected nested row committed, got count %d", count)
	}
	if count := countHelloDataByID(t, ctx, pool, joinedID); count != 1 {
		t.Fatalf("expected joined row committed, got count %d", count)
	}
}

func TestUnitOfWorkFactory_WithinTx_RequiresNew_CallbackPanic_RollsBackAndRePanics(t *testing.T) {
	pool := integration.SetupTestDB(t)
	defer integration.CleanupTestDB(t, pool)

	ctx := context.Background()
	factory := NewUnitOfWorkFactory(&PgxUnitOfWorkConfig{Pool: pool})
	testID := fmt.Sprintf("requires-new-panic-%d", time.Now().UnixNano())
	expectedPanic := "owned tx panic"

	var recovered any
	func() {
		defer func() {
			recovered = recover()
		}()

		_ = factory.WithinTx(ctx, transaction.RequiresNew(), func(txCtx context.Context, uow transaction.UnitOfWork) error {
			if _, err := uow.HelloRepository().Save(txCtx, &models.HelloData{Id: testID, Name: "PanicOwned", Age: 27}); err != nil {
				t.Fatalf("failed to save data before panic: %v", err)
			}
			panic(expectedPanic)
		})
	}()

	if recovered != expectedPanic {
		t.Fatalf("expected panic %q, got: %#v", expectedPanic, recovered)
	}
	if count := countHelloDataByID(t, ctx, pool, testID); count != 0 {
		t.Fatalf("expected panic path to rollback owned tx changes, got count %d", count)
	}
}

func TestUnitOfWorkFactory_WithinTx_Nested_CallbackPanic_RollsBackSavepointAndRePanics(t *testing.T) {
	pool := integration.SetupTestDB(t)
	defer integration.CleanupTestDB(t, pool)

	ctx := context.Background()
	factory := NewUnitOfWorkFactory(&PgxUnitOfWorkConfig{Pool: pool})

	outerID := fmt.Sprintf("nested-panic-outer-%d", time.Now().UnixNano())
	nestedID := fmt.Sprintf("nested-panic-inner-%d", time.Now().UnixNano())
	expectedPanic := "nested panic"

	var recovered any
	err := factory.WithinTx(ctx, transaction.RequiresNew(), func(txCtx context.Context, outer transaction.UnitOfWork) error {
		if _, saveErr := outer.HelloRepository().Save(txCtx, &models.HelloData{Id: outerID, Name: "OuterPanic", Age: 28}); saveErr != nil {
			return saveErr
		}

		func() {
			defer func() {
				recovered = recover()
			}()

			_ = factory.WithinTx(txCtx, transaction.Nested(), func(nestedCtx context.Context, nested transaction.UnitOfWork) error {
				if _, saveErr := nested.HelloRepository().Save(nestedCtx, &models.HelloData{Id: nestedID, Name: "InnerPanic", Age: 29}); saveErr != nil {
					return saveErr
				}
				panic(expectedPanic)
			})
		}()

		if recovered != expectedPanic {
			return fmt.Errorf("expected panic %q from nested flow, got: %#v", expectedPanic, recovered)
		}

		outerPgx, ok := outer.(*PgxUnitOfWork)
		if !ok {
			t.Fatalf("expected outer unit of work to be *PgxUnitOfWork")
		}
		if count := countHelloDataByIDTx(t, txCtx, outerPgx.txRef.current(), nestedID); count != 0 {
			t.Fatalf("expected nested panic path to rollback savepoint changes, got count %d", count)
		}

		return nil
	})
	if err != nil {
		t.Fatalf("expected outer tx to commit after recovering nested panic, got: %v", err)
	}

	if recovered != expectedPanic {
		t.Fatalf("expected panic %q, got: %#v", expectedPanic, recovered)
	}
	if count := countHelloDataByID(t, ctx, pool, outerID); count != 1 {
		t.Fatalf("expected outer row committed, got count %d", count)
	}
	if count := countHelloDataByID(t, ctx, pool, nestedID); count != 0 {
		t.Fatalf("expected nested panic row rolled back, got count %d", count)
	}
}

func TestSavepointTx_Commit_AfterCommit_ReturnsDeterministicError(t *testing.T) {
	tx := &savepointTx{savepoint: "sp_1", state: savepointCommitted}

	err := tx.Commit(context.Background())
	if !errors.Is(err, InvalidSavepointStateErr) {
		t.Fatalf("expected InvalidSavepointStateErr, got: %v", err)
	}
	if !strings.Contains(err.Error(), "cannot commit savepoint \"sp_1\" after commit") {
		t.Fatalf("expected deterministic commit savepoint error message, got: %v", err)
	}
}

func TestSavepointTx_Rollback_AfterRollback_ReturnsDeterministicError(t *testing.T) {
	tx := &savepointTx{savepoint: "sp_2", state: savepointRolledBack}

	err := tx.Rollback(context.Background())
	if !errors.Is(err, InvalidSavepointStateErr) {
		t.Fatalf("expected InvalidSavepointStateErr, got: %v", err)
	}
	if !strings.Contains(err.Error(), "cannot rollback savepoint \"sp_2\" after rollback") {
		t.Fatalf("expected deterministic rollback savepoint error message, got: %v", err)
	}
}

func countHelloDataByID(t *testing.T, ctx context.Context, pool *pgxpool.Pool, id string) int {
	t.Helper()

	var count int
	if err := pool.QueryRow(ctx, "SELECT COUNT(*) FROM HELLO_DATA WHERE ID = $1", id).Scan(&count); err != nil {
		t.Fatalf("failed to count rows for id %s: %v", id, err)
	}
	return count
}

func countHelloDataByIDTx(t *testing.T, ctx context.Context, tx pgx.Tx, id string) int {
	t.Helper()

	var count int
	if err := tx.QueryRow(ctx, "SELECT COUNT(*) FROM HELLO_DATA WHERE ID = $1", id).Scan(&count); err != nil {
		t.Fatalf("failed to count rows in tx for id %s: %v", id, err)
	}
	return count
}
