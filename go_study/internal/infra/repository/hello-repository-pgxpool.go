package repository

import (
	"context"
	"main/internal/hello"
	"main/internal/infra/observability/trace"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	tracer = trace.GetTracer("HelloRepository")
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
	ctx, span := tracer.StartSpan(ctx, "Save")
	span.SetAttributes(trace.Attribute("id", entity.Id))
	defer span.End()
	const sql = "INSERT INTO HELLO_DATA (ID, NAME, AGE) VALUES ($1, $2, $3)"
	var err error = nil
	var params = []any{entity.Id, entity.Name, entity.Age}

	tx := r.getTransactionOrNil(ctx)
	if tx == nil {
		_, err = r.pool.Exec(ctx, sql, params...)
	} else {
		_, err = (*tx).Exec(ctx, sql, params...)
	}

	if err != nil {
		span.SetError(err)
		return nil, err
	}
	return entity, nil
}

func (r HelloRepository) FindById(ctx context.Context, id any) (*hello.HelloData, error) {
	ctx, span := tracer.StartSpan(ctx, "FindById")
	defer span.End()
	const sql = "SELECT ID, NAME, AGE FROM HELLO_DATA WHERE ID = $1"
	strId := id.(string)
	span.SetAttributes(trace.Attribute("id", strId))

	tx := r.getTransactionOrNil(ctx)

	if tx == nil {
		row := r.pool.QueryRow(ctx, sql, strId)
		return r.mapRow(ctx, &row)
	} else {
		row := (*tx).QueryRow(ctx, sql, strId)
		return r.mapRow(ctx, &row)
	}
}

func (r HelloRepository) ListAll(ctx context.Context) []hello.HelloData {
	ctx, span := tracer.StartSpan(ctx, "ListAll")
	defer span.End()
	const sql = "SELECT ID, NAME, AGE FROM HELLO_DATA"
	var rows pgx.Rows = nil
	var err error = nil
	tx := r.getTransactionOrNil(ctx)
	if tx == nil {
		rows, err = r.pool.Query(ctx, sql)
	} else {
		rows, err = (*tx).Query(ctx, sql)
	}

	if err != nil {
		span.SetError(err)
		return nil
	}

	result := make([]hello.HelloData, 0)

	for rows.Next() {
		data, err := r.mapRowsNextValue(ctx, &rows)
		if err != nil {
			span.SetError(err)
			return make([]hello.HelloData, 0)
		}
		result = append(result, *data)
	}
	return result
}

func (r HelloRepository) BeginTransactionWithContext(ctx context.Context) (context.Context, error) {
	ctx, span := tracer.StartSpan(ctx, "BeginTransactionWithContext")
	defer span.End()
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		span.SetError(err)
		return nil, err
	}
	return context.WithValue(ctx, txKey, &tx), nil
}

func (r HelloRepository) Rollback(ctx context.Context) {
	ctx, span := tracer.StartSpan(ctx, "Rollback")
	defer span.End()
	tx := r.getTransactionOrNil(ctx)
	if tx != nil {
		(*tx).Rollback(ctx)
	}
}

func (r HelloRepository) Commit(ctx context.Context) {
	ctx, span := tracer.StartSpan(ctx, "Commit")
	defer span.End()
	tx := r.getTransactionOrNil(ctx)
	if tx != nil {
		(*tx).Commit(ctx)
	}
}

func (r HelloRepository) getTransactionOrNil(ctx context.Context) *pgx.Tx {
	ctx, span := tracer.StartSpan(ctx, "getTransactionOrNil")
	defer span.End()
	tx := ctx.Value(txKey)
	if tx == nil {
		return nil
	}
	return tx.(*pgx.Tx)
}

func (r HelloRepository) RunWithTransaction(ctx context.Context, callback func(context.Context) (*hello.HelloData, error)) (*hello.HelloData, error) {
	ctx, span := tracer.StartSpan(ctx, "RunWithTransaction")
	defer span.End()
	ctxTx, err := r.BeginTransactionWithContext(ctx)
	if err != nil {
		span.SetError(err)
		return nil, err
	}
	result, err := callback(ctxTx)
	if err != nil {
		span.SetError(err)
		r.Rollback(ctxTx)
	} else {
		r.Commit(ctxTx)
	}
	return result, err
}

func (r HelloRepository) mapRow(ctx context.Context, row *pgx.Row) (*hello.HelloData, error) {
	_, span := tracer.StartSpan(ctx, "mapRow")
	defer span.End()
	result := new(hello.HelloData)
	err := (*row).Scan(&result.Id, &result.Name, &result.Age)
	if err != nil {
		span.SetError(err)
	}
	return result, err
}

func (r HelloRepository) mapRowsNextValue(ctx context.Context, rows *pgx.Rows) (*hello.HelloData, error) {
	_, span := tracer.StartSpan(ctx, "mapRowsNextValue")
	defer span.End()
	result := new(hello.HelloData)
	err := (*rows).Scan(&result.Id, &result.Name, &result.Age)
	if err != nil {
		span.SetError(err)
	}
	return result, err
}
