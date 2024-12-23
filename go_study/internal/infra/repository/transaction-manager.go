package repository

import (
	"context"

	"github.com/bruno303/study-topics/go-study/internal/crosscutting/observability/trace"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type (
	TransactionManager struct {
		config *TransactionConfig
	}
	ApplicationTransaction struct {
		postgreTransaction *pgx.Tx
		fileTransaction    *fileTransaction
	}
	TransactionConfig struct {
		Pool *pgxpool.Pool
	}
	txCtxKeyType string
)

var txCtxKey txCtxKeyType = txCtxKeyType("transaction-key")

func newFileTransaction() *fileTransaction {
	return &fileTransaction{
		paths: make(map[string]string),
	}
}

func NewTransactionManager(cfg *TransactionConfig) TransactionManager {
	return TransactionManager{cfg}
}

func (tm TransactionManager) Execute(ctx context.Context, callback func(txCtx context.Context) (any, error)) (any, error) {
	var result any
	ctx, end := trace.Trace(ctx, trace.NameConfig("TransactionManager", "Execute"))
	defer end()

	ctx, err := tm.BeginTransaction(ctx)
	if err != nil {
		trace.InjectError(ctx, err)
		return result, err
	}
	cbResult, err := callback(ctx)
	tx := GetTransactionOrNil(ctx)

	if tx != nil {
		if err != nil {
			err = tx.Rollback(ctx)
		} else {
			err = tx.Commit(ctx)
		}
	}

	if err != nil {
		trace.InjectError(ctx, err)
		return result, err
	}
	return cbResult, nil
}

func (tm TransactionManager) BeginTransaction(ctx context.Context) (context.Context, error) {
	ctx, end := trace.Trace(ctx, trace.NameConfig("TransactionManager", "BeginTransaction"))
	defer end()

	tx, err := tm.config.Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		trace.InjectError(ctx, err)
		return ctx, err
	}
	return context.WithValue(ctx, txCtxKey, &ApplicationTransaction{&tx, newFileTransaction()}), nil
}

func GetTransactionOrNil(ctx context.Context) *ApplicationTransaction {
	value := ctx.Value(txCtxKey)
	if appTx, ok := value.(*ApplicationTransaction); ok {
		return appTx
	}
	return nil
}

func (t *ApplicationTransaction) Commit(ctx context.Context) error {
	ctx, end := trace.Trace(ctx, trace.NameConfig("TransactionManager", "Commit"))
	defer end()
	if t.fileTransaction != nil {
		if err := t.fileTransaction.Commit(ctx); err != nil {
			trace.InjectError(ctx, err)
			return err
		}
	}
	if t.postgreTransaction != nil {
		if err := (*t.postgreTransaction).Commit(ctx); err != nil {
			trace.InjectError(ctx, err)
			return err
		}
	}
	return nil
}

func (t *ApplicationTransaction) Rollback(ctx context.Context) error {
	ctx, end := trace.Trace(ctx, trace.NameConfig("TransactionManager", "Rollback"))
	defer end()
	if t.fileTransaction != nil {
		if err := t.fileTransaction.Rollback(ctx); err != nil {
			trace.InjectError(ctx, err)
			return err
		}
	}
	if t.postgreTransaction != nil {
		if err := (*t.postgreTransaction).Rollback(ctx); err != nil {
			trace.InjectError(ctx, err)
			return err
		}
	}
	return nil
}
