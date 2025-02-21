package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OptimizedTransaction struct {
	pool       *pgxpool.Pool
	operations []func(tx *pgx.Tx) error
}

func NewOptimizedTransaction(pool *pgxpool.Pool) OptimizedTransaction {
	return OptimizedTransaction{
		pool:       pool,
		operations: make([]func(tx *pgx.Tx) error, 0),
	}
}

func (t *OptimizedTransaction) AddOperation(operation func(tx *pgx.Tx) error) {
	t.operations = append(t.operations, operation)
}

func (t *OptimizedTransaction) Commit(ctx context.Context) error {
	tx, err := t.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	for _, operation := range t.operations {
		if err := operation(&tx); err != nil {
			tx.Rollback(ctx)
			return err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

type OptimizedTransactionManager struct {
	pool *pgxpool.Pool
}

func NewOptimizedTransactionManager(pool *pgxpool.Pool) OptimizedTransactionManager {
	return OptimizedTransactionManager{
		pool: pool,
	}
}

func (tm OptimizedTransactionManager) Execute(ctx context.Context, callback func(context.Context) (any, error)) (any, error) {
	tx := NewOptimizedTransaction(tm.pool)
	ctx = context.WithValue(ctx, txCtxKey, &tx)

	cbResult, err := callback(ctx)
	if err != nil {
		return nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}
	return cbResult, nil
}
