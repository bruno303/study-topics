package repository

import (
	"context"

	"github.com/bruno303/study-topics/go-study/internal/crosscutting/observability/trace"
	"github.com/bruno303/study-topics/go-study/internal/crosscutting/observability/trace/attr"
	"github.com/bruno303/study-topics/go-study/internal/hello"

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

const traceName = "HelloRepository"

func NewHelloPgxRepository(pool *pgxpool.Pool) HelloRepository {
	return HelloRepository{
		pool: pool,
	}
}

func (r HelloRepository) Save(ctx context.Context, entity *hello.HelloData) (*hello.HelloData, error) {
	ctx, end := trace.Trace(ctx, trace.NameConfig(traceName, "Save"))
	defer end()

	trace.InjectAttributes(ctx, attr.New("id", entity.Id))

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
		trace.InjectError(ctx, err)
		return nil, err
	}
	return entity, nil
}

func (r HelloRepository) FindById(ctx context.Context, id any) (*hello.HelloData, error) {
	ctx, end := trace.Trace(ctx, trace.NameConfig(traceName, "FindById"))
	defer end()

	strId := id.(string)
	trace.InjectAttributes(ctx, attr.New("id", strId))
	const sql = "SELECT ID, NAME, AGE FROM HELLO_DATA WHERE ID = $1"
	tx := r.getTransactionOrNil(ctx)

	if tx == nil {
		row := r.pool.QueryRow(ctx, sql, strId)
		return r.mapRow(&row)
	} else {
		row := (*tx).QueryRow(ctx, sql, strId)
		return r.mapRow(&row)
	}
}

func (r HelloRepository) ListAll(ctx context.Context) []hello.HelloData {
	ctx, end := trace.Trace(ctx, trace.NameConfig(traceName, "ListAll"))
	defer end()

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
		trace.InjectError(ctx, err)
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
}

func (r HelloRepository) BeginTransactionWithContext(ctx context.Context) (context.Context, error) {
	// save context in a new variable to avoid returning this span up in the chain
	sCtx, end := trace.Trace(ctx, trace.NameConfig(traceName, "BeginTransactionWithContext"))
	defer end()

	tx, err := r.pool.BeginTx(sCtx, pgx.TxOptions{})
	if err != nil {
		trace.InjectError(sCtx, err)
		return nil, err
	}
	return context.WithValue(ctx, txKey, &tx), nil
}

func (r HelloRepository) getTransactionOrNil(ctx context.Context) *pgx.Tx {
	appTx := GetTransactionOrNil(ctx)
	if appTx != nil {
		return appTx.postgreTransaction
	}
	return nil
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
