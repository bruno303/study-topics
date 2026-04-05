package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/bruno303/study-topics/go-study/internal/application/repository"
	"github.com/bruno303/study-topics/go-study/internal/crosscutting/observability/trace"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type (
	transactionRef struct {
		tx pgx.Tx
	}
	PgxUnitOfWork struct {
		config          *PgxUnitOfWorkConfig
		txRef           *transactionRef
		helloRepository repository.HelloRepository
	}
	PgxUnitOfWorkConfig struct {
		Pool *pgxpool.Pool
	}
)

var (
	InvalidTransactionTypeErr = errors.New("invalid transaction type")
	InvalidPropagationErr     = errors.New("invalid propagation option")
	TransactionAlreadyOpenErr = errors.New("transaction already opened")
	TransactionNotOpenedErr   = errors.New("transaction not opened")
)

func NewPgxUnitOfWork(cfg *PgxUnitOfWorkConfig) *PgxUnitOfWork {
	return newPgxUnitOfWorkWithTxRef(cfg, &transactionRef{})
}

func newPgxUnitOfWorkWithTxRef(cfg *PgxUnitOfWorkConfig, txRef *transactionRef) *PgxUnitOfWork {
	return &PgxUnitOfWork{
		config:          cfg,
		txRef:           txRef,
		helloRepository: newHelloPgxRepository(cfg.Pool, txRef),
	}
}

func (tm *PgxUnitOfWork) HelloRepository() repository.HelloRepository {
	return tm.helloRepository
}

func (tm *PgxUnitOfWork) Begin(ctx context.Context) error {
	ctx, end := trace.Trace(ctx, trace.NameConfig("PgxUnitOfWork", "Begin"))
	defer end()

	if tm.txRef.current() != nil {
		err := fmt.Errorf("%w: begin is not allowed for externally scoped transactions", TransactionAlreadyOpenErr)
		trace.InjectError(ctx, err)
		return err
	}

	tx, err := tm.config.Pool.Begin(ctx)
	if err != nil {
		trace.InjectError(ctx, err)
		return err
	}

	tm.txRef.set(tx)
	return nil
}

func (uow *PgxUnitOfWork) Commit(ctx context.Context) error {
	if uow.txRef.current() == nil {
		return TransactionNotOpenedErr
	}
	defer uow.clear()
	return uow.txRef.current().Commit(ctx)
}

func (uow *PgxUnitOfWork) Rollback(ctx context.Context) error {
	if uow.txRef.current() == nil {
		return TransactionNotOpenedErr
	}
	defer uow.clear()
	return uow.txRef.current().Rollback(ctx)
}

func (tm *PgxUnitOfWork) clear() {
	tm.txRef.set(nil)
}

func (r *transactionRef) current() pgx.Tx {
	if r == nil {
		return nil
	}
	return r.tx
}

func (r *transactionRef) set(tx pgx.Tx) {
	if r == nil {
		return
	}
	r.tx = tx
}
