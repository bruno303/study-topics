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
	pgxRepositoryAccessor struct {
		helloRepository applicationRepository.HelloRepository
	}

	PgxUnitOfWork struct {
		config *PgxUnitOfWorkConfig
	}

	PgxUnitOfWorkConfig struct {
		Pool *pgxpool.Pool
	}
)

var _ transaction.UnitOfWork = (*PgxUnitOfWork)(nil)
var _ transaction.RepositoryAccessor = (*pgxRepositoryAccessor)(nil)

func NewPgxUnitOfWork(cfg *PgxUnitOfWorkConfig) *PgxUnitOfWork {
	return &PgxUnitOfWork{config: cfg}
}

func (uow *PgxUnitOfWork) WithinTx(ctx context.Context, fn transaction.TransactionCallback) error {
	ctx, end := trace.Trace(ctx, trace.NameConfig("PgxUnitOfWork", "WithinTx"))
	defer end()

	tx, err := uow.config.Pool.Begin(ctx)
	if err != nil {
		trace.InjectError(ctx, err)
		return err
	}

	repos := uow.newRepositoryAccessor(tx)
	return uow.execute(ctx, tx, repos, fn)
}

func (uow *PgxUnitOfWork) newRepositoryAccessor(tx pgx.Tx) transaction.RepositoryAccessor {
	return &pgxRepositoryAccessor{
		helloRepository: newHelloPgxRepository(uow.config.Pool, tx),
	}
}

func (uow *PgxUnitOfWork) execute(ctx context.Context, tx pgx.Tx, repos transaction.RepositoryAccessor, fn transaction.TransactionCallback) error {
	var callbackPanic any
	var callbackErr error
	func() {
		defer func() {
			if recovered := recover(); recovered != nil {
				callbackPanic = recovered
			}
		}()

		callbackErr = fn(ctx, repos)
	}()

	if callbackPanic != nil {
		uow.rollbackBestEffort(ctx, tx)
		panic(callbackPanic)
	}

	if callbackErr != nil {
		rollbackErr := uow.rollback(ctx, tx)
		if rollbackErr != nil {
			joinedErr := combineCallbackAndRollbackErr(callbackErr, rollbackErr)
			trace.InjectError(ctx, joinedErr)
			return joinedErr
		}
		return callbackErr
	}

	commitErr := uow.commit(ctx, tx)
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

func (uow *PgxUnitOfWork) commit(ctx context.Context, tx pgx.Tx) error {
	ctx, end := trace.Trace(ctx, trace.NameConfig("PgxUnitOfWork", "Commit"))
	defer end()

	err := tx.Commit(ctx)
	if err != nil {
		trace.InjectError(ctx, err)
		return err
	}

	return nil
}

func (uow *PgxUnitOfWork) rollback(ctx context.Context, tx pgx.Tx) error {
	ctx, end := trace.Trace(ctx, trace.NameConfig("PgxUnitOfWork", "Rollback"))
	defer end()

	err := tx.Rollback(ctx)
	if err != nil {
		trace.InjectError(ctx, err)
		return err
	}

	return nil
}

func (uow *PgxUnitOfWork) rollbackBestEffort(ctx context.Context, tx pgx.Tx) {
	defer func() {
		if recovered := recover(); recovered != nil {
			trace.InjectError(ctx, fmt.Errorf("panic during rollback: %v", recovered))
		}
	}()

	rollbackErr := uow.rollback(ctx, tx)
	if rollbackErr != nil {
		trace.InjectError(ctx, rollbackErr)
	}
}

func (a *pgxRepositoryAccessor) HelloRepository() applicationRepository.HelloRepository {
	return a.helloRepository
}
