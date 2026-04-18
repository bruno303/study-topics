package repository

import (
	"context"
	"errors"
	"fmt"

	applicationRepository "github.com/bruno303/study-topics/go-study/internal/application/repository"
	"github.com/bruno303/study-topics/go-study/internal/application/transaction"
	"github.com/bruno303/study-topics/go-study/internal/crosscutting/observability/trace"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type (
	pgxUnitOfWork struct {
		helloRepository  applicationRepository.HelloRepository
		outboxRepository applicationRepository.OutboxRepository
	}

	PgxTransactionManager struct {
		config *PgxTransactionManagerConfig
	}

	PgxTransactionManagerConfig struct {
		Pool *pgxpool.Pool
	}
)

var _ transaction.TransactionManager = (*PgxTransactionManager)(nil)
var _ transaction.UnitOfWork = (*pgxUnitOfWork)(nil)

func NewPgxTransactionManager(cfg *PgxTransactionManagerConfig) *PgxTransactionManager {
	return &PgxTransactionManager{config: cfg}
}

func (tm *PgxTransactionManager) WithinTx(ctx context.Context, opts transaction.TransactionOpts, fn transaction.TransactionCallback) error {
	ctx, end := trace.Trace(ctx, trace.NameConfig("PgxTransactionManager", "WithinTx"))
	defer end()

	if opts.Parent != nil {
		return fn(ctx, opts.Parent)
	}

	tx, err := tm.config.Pool.Begin(ctx)
	if err != nil {
		trace.InjectError(ctx, err)
		return err
	}

	uow := tm.newUnitOfWork(tx)
	return tm.execute(ctx, tx, uow, fn)
}

func (tm *PgxTransactionManager) newUnitOfWork(tx pgx.Tx) transaction.UnitOfWork {
	return &pgxUnitOfWork{
		helloRepository:  newHelloPgxRepository(tm.config.Pool, tx),
		outboxRepository: NewOutboxPgxRepository(tm.config.Pool, tx),
	}
}

func (tm *PgxTransactionManager) execute(ctx context.Context, tx pgx.Tx, uow transaction.UnitOfWork, fn transaction.TransactionCallback) error {
	var callbackPanic any
	var callbackErr error
	func() {
		defer func() {
			if recovered := recover(); recovered != nil {
				callbackPanic = recovered
			}
		}()

		callbackErr = fn(ctx, uow)
	}()

	if callbackPanic != nil {
		tm.rollbackBestEffort(ctx, tx)
		panic(callbackPanic)
	}

	if callbackErr != nil {
		rollbackErr := tm.rollback(ctx, tx)
		if rollbackErr != nil {
			joinedErr := combineCallbackAndRollbackErr(callbackErr, rollbackErr)
			trace.InjectError(ctx, joinedErr)
			return joinedErr
		}
		return callbackErr
	}

	commitErr := tm.commit(ctx, tx)
	if commitErr != nil {
		trace.InjectError(ctx, commitErr)
		return commitErr
	}

	return nil
}

func combineCallbackAndRollbackErr(callbackErr, rollbackErr error) error {
	if rollbackErr == nil {
		return callbackErr
	}
	return errors.Join(callbackErr, fmt.Errorf("rollback transaction: %w", rollbackErr))
}

func (tm *PgxTransactionManager) commit(ctx context.Context, tx pgx.Tx) error {
	ctx, end := trace.Trace(ctx, trace.NameConfig("PgxTransactionManager", "Commit"))
	defer end()

	err := tx.Commit(ctx)
	if err != nil {
		trace.InjectError(ctx, err)
		return err
	}

	return nil
}

func (tm *PgxTransactionManager) rollback(ctx context.Context, tx pgx.Tx) error {
	ctx, end := trace.Trace(ctx, trace.NameConfig("PgxTransactionManager", "Rollback"))
	defer end()

	err := tx.Rollback(ctx)
	if err != nil {
		trace.InjectError(ctx, err)
		return err
	}

	return nil
}

func (tm *PgxTransactionManager) rollbackBestEffort(ctx context.Context, tx pgx.Tx) {
	defer func() {
		if recovered := recover(); recovered != nil {
			trace.InjectError(ctx, fmt.Errorf("panic during rollback: %v", recovered))
		}
	}()

	rollbackErr := tm.rollback(ctx, tx)
	if rollbackErr != nil {
		trace.InjectError(ctx, rollbackErr)
	}
}

func (uow *pgxUnitOfWork) HelloRepository() applicationRepository.HelloRepository {
	return uow.helloRepository
}

func (uow *pgxUnitOfWork) OutboxRepository() applicationRepository.OutboxRepository {
	return uow.outboxRepository
}
