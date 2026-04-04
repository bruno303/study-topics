package repository

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/bruno303/study-topics/go-study/internal/application/transaction"
	"github.com/bruno303/study-topics/go-study/internal/crosscutting/observability/trace"
	"github.com/jackc/pgx/v5"
)

type UnitOfWorkFactory struct {
	config *PgxUnitOfWorkConfig
}

type txScopeContextKey struct{}

type txScope struct {
	tx pgx.Tx
}

var savepointCounter atomic.Uint64

var _ transaction.TransactionManager = (*UnitOfWorkFactory)(nil)

func NewUnitOfWorkFactory(config *PgxUnitOfWorkConfig) UnitOfWorkFactory {
	return UnitOfWorkFactory{config: config}
}

func (f UnitOfWorkFactory) New() transaction.UnitOfWork {
	return NewPgxUnitOfWork(f.config)
}

func (f UnitOfWorkFactory) WithinTx(ctx context.Context, opts transaction.Opts, fn func(context.Context, transaction.UnitOfWork) error) error {
	ctx, end := trace.Trace(ctx, trace.NameConfig("UnitOfWorkFactory", "WithinTx"))
	defer end()

	parent := currentTxScope(ctx)
	propagation := opts.EffectivePropagation()

	switch propagation {
	case transaction.PropagationJoin:
		if parent == nil {
			return f.withOwnTransaction(ctx, fn)
		}
		uow := f.newScopedUnitOfWork(parent.tx, joinedScope)
		return fn(ctx, uow)
	case transaction.PropagationRequiresNew:
		return f.withOwnTransaction(ctx, fn)
	case transaction.PropagationNested:
		if parent == nil {
			return f.withOwnTransaction(ctx, fn)
		}
		return f.withSavepoint(ctx, parent.tx, fn)
	default:
		trace.InjectError(ctx, InvalidPropagationErr)
		return InvalidPropagationErr
	}
}

func (f UnitOfWorkFactory) withOwnTransaction(ctx context.Context, fn func(context.Context, transaction.UnitOfWork) error) error {
	uow, err := f.newOwnedUnitOfWork(ctx)
	if err != nil {
		trace.InjectError(ctx, err)
		return err
	}

	pgxUow, err := asPgxUnitOfWork(uow)
	if err != nil {
		trace.InjectError(ctx, err)
		return err
	}

	scopedCtx := withTxScope(ctx, pgxUow.txRef.current())
	var callbackPanic any
	var callbackErr error
	func() {
		defer func() {
			if recovered := recover(); recovered != nil {
				callbackPanic = recovered
			}
		}()

		callbackErr = fn(scopedCtx, uow)
	}()

	if callbackPanic != nil {
		rollbackBestEffort(scopedCtx, uow)
		panic(callbackPanic)
	}

	if callbackErr != nil {
		rollbackErr := uow.Rollback(scopedCtx)
		if rollbackErr != nil {
			trace.InjectError(scopedCtx, rollbackErr)
			return rollbackErr
		}
		return callbackErr
	}

	commitErr := uow.Commit(scopedCtx)
	if commitErr != nil {
		trace.InjectError(scopedCtx, commitErr)
		return commitErr
	}

	return nil
}

func (f UnitOfWorkFactory) withSavepoint(ctx context.Context, parentTx pgx.Tx, fn func(context.Context, transaction.UnitOfWork) error) error {
	savepointName := fmt.Sprintf("sp_%d", savepointCounter.Add(1))
	_, err := parentTx.Exec(ctx, fmt.Sprintf("SAVEPOINT %s", savepointName))
	if err != nil {
		trace.InjectError(ctx, err)
		return err
	}

	nestedTx := newSavepointTx(parentTx, savepointName)
	uow := f.newScopedUnitOfWork(nestedTx, nestedScope)
	scopedCtx := withTxScope(ctx, nestedTx)

	var callbackPanic any
	var callbackErr error
	func() {
		defer func() {
			if recovered := recover(); recovered != nil {
				callbackPanic = recovered
			}
		}()

		callbackErr = fn(scopedCtx, uow)
	}()

	if callbackPanic != nil {
		rollbackBestEffort(scopedCtx, uow)
		panic(callbackPanic)
	}

	if callbackErr != nil {
		rollbackErr := uow.Rollback(scopedCtx)
		if rollbackErr != nil {
			trace.InjectError(scopedCtx, rollbackErr)
			return rollbackErr
		}
		return callbackErr
	}

	commitErr := uow.Commit(scopedCtx)
	if commitErr != nil {
		trace.InjectError(scopedCtx, commitErr)
		return commitErr
	}

	return nil
}

func (f UnitOfWorkFactory) newOwnedUnitOfWork(ctx context.Context) (transaction.UnitOfWork, error) {
	uow := f.New()
	err := uow.Begin(ctx, transaction.RequiresNew())
	if err != nil {
		return nil, err
	}
	return uow, nil
}

func (f UnitOfWorkFactory) newScopedUnitOfWork(tx pgx.Tx, scope transactionScope) transaction.UnitOfWork {
	txRef := &transactionRef{}
	txRef.set(tx)
	return newPgxUnitOfWorkWithTxRef(f.config, txRef, scope)
}

func asPgxUnitOfWork(uow transaction.UnitOfWork) (*PgxUnitOfWork, error) {
	pgxUow, ok := uow.(*PgxUnitOfWork)
	if !ok {
		return nil, InvalidTransactionTypeErr
	}
	return pgxUow, nil
}

func currentTxScope(ctx context.Context) *txScope {
	scope, ok := ctx.Value(txScopeContextKey{}).(*txScope)
	if !ok {
		return nil
	}
	return scope
}

func withTxScope(ctx context.Context, tx pgx.Tx) context.Context {
	if tx == nil {
		return ctx
	}
	return context.WithValue(ctx, txScopeContextKey{}, &txScope{tx: tx})
}

func rollbackBestEffort(ctx context.Context, uow transaction.UnitOfWork) {
	defer func() {
		if recovered := recover(); recovered != nil {
			trace.InjectError(ctx, fmt.Errorf("panic during rollback: %v", recovered))
		}
	}()

	rollbackErr := uow.Rollback(ctx)
	if rollbackErr != nil {
		trace.InjectError(ctx, rollbackErr)
	}
}
