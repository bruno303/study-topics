package repository

import (
	"context"
	"errors"

	"github.com/bruno303/study-topics/go-study/internal/application/transaction"
	"github.com/bruno303/study-topics/go-study/internal/crosscutting/observability/trace"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type (
	TransactionManager struct {
		config *TransactionConfig
	}
	TransactionConfig struct {
		Pool *pgxpool.Pool
	}
	txCtxKeyType string
)

var (
	_ transaction.TransactionManager = (*TransactionManager)(nil)
)

var (
	txCtxKey                  txCtxKeyType = txCtxKeyType("transaction-key")
	InvalidTransactionTypeErr              = errors.New("invalid transaction type")
)

func NewTransactionManager(cfg *TransactionConfig) TransactionManager {
	return TransactionManager{cfg}
}

func (tm TransactionManager) Execute(ctx context.Context, opts transaction.Opts, callback transaction.TransactionalFunc) (any, error) {
	ctx, end := trace.Trace(ctx, trace.NameConfig("TransactionManager", "Execute"))
	defer end()

	var (
		empty any
		tx    pgx.Tx
		err   error
	)

	// If RequiresNew is true, always create a new transaction
	// If RequiresNew is false and a transaction is provided, use it
	// Otherwise, create a new transaction
	if opts.RequiresNew {
		tx, err = tm.beginTransaction(ctx)
		if err != nil {
			trace.InjectError(ctx, err)
			return empty, err
		}
	} else if opts.Transaction != nil {
		var ok bool
		tx, ok = opts.Transaction.(pgx.Tx)
		if !ok {
			trace.InjectError(ctx, InvalidTransactionTypeErr)
			return empty, InvalidTransactionTypeErr
		}
	} else {
		tx, err = tm.beginTransaction(ctx)
		if err != nil {
			trace.InjectError(ctx, err)
			return empty, err
		}
	}

	ctx = context.WithValue(ctx, txCtxKey, tx)
	cbResult, err := callback(ctx, tx)

	if err != nil {
		trace.InjectError(ctx, err)

		txErr := tx.Rollback(ctx)
		if txErr != nil {
			trace.InjectError(ctx, txErr)
		}
		return empty, err
	}

	txErr := tx.Commit(ctx)
	if txErr != nil {
		trace.InjectError(ctx, txErr)
		return empty, txErr
	}

	return cbResult, nil
}

func (tm TransactionManager) beginTransaction(ctx context.Context) (pgx.Tx, error) {
	ctx, end := trace.Trace(ctx, trace.NameConfig("TransactionManager", "BeginTransaction"))
	defer end()

	return tm.config.Pool.Begin(ctx)
}
