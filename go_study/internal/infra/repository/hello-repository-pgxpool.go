package repository

import (
	"context"
	"main/internal/crosscutting/observability"
	"main/internal/hello"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type HelloRepository struct {
	pool *pgxpool.Pool
}

type transactionKey struct {
	name string
}

var txKey = transactionKey{name: "db-transaction"}

func NewHelloRepository(ctx context.Context, pool *pgxpool.Pool) HelloRepository {
	return HelloRepository{
		pool: pool,
	}
}

func (r HelloRepository) Save(ctx context.Context, entity *hello.HelloData) (*hello.HelloData, error) {
	const sql = "INSERT INTO HELLO_DATA (ID, NAME, AGE) VALUES ($1, $2, $3)"
	return observability.WithTracingBiResult(ctx, "HelloRepository", "Save", func(ctx context.Context) (*hello.HelloData, error) {
		var err error = nil
		var params = []any{entity.Id, entity.Name, entity.Age}

		tx := r.getTransactionOrNil(ctx)
		if tx == nil {
			_, err = r.pool.Exec(ctx, sql, params...)
		} else {
			_, err = (*tx).Exec(ctx, sql, params...)
		}

		if err != nil {
			return nil, err
		}
		return entity, nil
	})
}

func (r HelloRepository) FindById(ctx context.Context, id any) (*hello.HelloData, error) {
	const sql = "SELECT ID, NAME, AGE FROM HELLO_DATA WHERE ID = $1"
	return observability.WithTracingBiResult(ctx, "HelloRepository", "FindById", func(ctx context.Context) (*hello.HelloData, error) {
		strId := id.(string)
		tx := r.getTransactionOrNil(ctx)

		if tx == nil {
			row := r.pool.QueryRow(ctx, sql, strId)
			return r.mapRow(&row)
		} else {
			row := (*tx).QueryRow(ctx, sql, strId)
			return r.mapRow(&row)
		}
	})
}

func (r HelloRepository) ListAll(ctx context.Context) []hello.HelloData {
	const sql = "SELECT ID, NAME, AGE FROM HELLO_DATA"
	return observability.WithTracingResult(ctx, "HelloRepository", "FindById", func(ctx context.Context) []hello.HelloData {
		var rows pgx.Rows = nil
		var err error = nil
		tx := r.getTransactionOrNil(ctx)
		if tx == nil {
			rows, err = r.pool.Query(ctx, sql)
		} else {
			rows, err = (*tx).Query(ctx, sql)
		}

		if err != nil {
			return nil
		}

		result := make([]hello.HelloData, 0)

		for rows.Next() {
			data, err := r.mapRowsNextValue(&rows)
			if err != nil {
				return make([]hello.HelloData, 0)
			}
			result = append(result, *data)
		}
		return result
	})
}

func (r HelloRepository) BeginTransactionWithContext(ctx context.Context) (context.Context, error) {
	return observability.WithTracingBiResult(ctx, "HelloRepository", "BeginTransactionWithContext", func(ctx context.Context) (context.Context, error) {
		tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			return nil, err
		}
		return context.WithValue(ctx, txKey, &tx), nil
	})
}

func (r HelloRepository) Rollback(ctx context.Context) {
	observability.WithTracing(ctx, "HelloRepository", "Rollback", func(ctx context.Context) {
		tx := r.getTransactionOrNil(ctx)
		if tx != nil {
			(*tx).Rollback(ctx)
		}
	})
}

func (r HelloRepository) Commit(ctx context.Context) {
	observability.WithTracing(ctx, "HelloRepository", "Commit", func(ctx context.Context) {
		tx := r.getTransactionOrNil(ctx)
		if tx != nil {
			(*tx).Commit(ctx)
		}
	})
}

func (r HelloRepository) getTransactionOrNil(ctx context.Context) *pgx.Tx {
	tx := ctx.Value(txKey)
	if tx == nil {
		return nil
	}
	return tx.(*pgx.Tx)
}

func (r HelloRepository) RunWithTransaction(ctx context.Context, callback func(context.Context) (*hello.HelloData, error)) (*hello.HelloData, error) {
	return observability.WithTracingBiResult(ctx, "HelloRepository", "RunWithTransaction", func(ctx context.Context) (*hello.HelloData, error) {
		ctxTx, err := r.BeginTransactionWithContext(ctx)
		if err != nil {
			return nil, err
		}
		result, err := callback(ctxTx)
		if err != nil {
			r.Rollback(ctxTx)
		} else {
			r.Commit(ctxTx)
		}
		return result, err
	})
}

func (r HelloRepository) mapRow(row *pgx.Row) (*hello.HelloData, error) {
	result := new(hello.HelloData)
	err := (*row).Scan(&result.Id, &result.Name, &result.Age)
	return result, err
}

func (r HelloRepository) mapRowsNextValue(rows *pgx.Rows) (*hello.HelloData, error) {
	result := new(hello.HelloData)
	err := (*rows).Scan(&result.Id, &result.Name, &result.Age)
	return result, err
}
