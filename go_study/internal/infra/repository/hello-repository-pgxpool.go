package repository

import (
	"context"
	"fmt"

	"github.com/bruno303/study-topics/go-study/internal/application/hello/models"
	"github.com/bruno303/study-topics/go-study/internal/crosscutting/observability/trace"
	"github.com/bruno303/study-topics/go-study/internal/crosscutting/observability/trace/attr"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type HelloRepository struct {
	pool  *pgxpool.Pool
	txRef *transactionRef
}

const traceName = "HelloRepository"

func NewHelloPgxRepository(pool *pgxpool.Pool) HelloRepository {
	return newHelloPgxRepository(pool, &transactionRef{})
}

func newHelloPgxRepository(pool *pgxpool.Pool, txRef *transactionRef) HelloRepository {
	return HelloRepository{
		pool:  pool,
		txRef: txRef,
	}
}

func (r HelloRepository) Save(ctx context.Context, entity *models.HelloData) (*models.HelloData, error) {
	ctx, end := trace.Trace(ctx, trace.NameConfig(traceName, "Save"))
	defer end()

	trace.InjectAttributes(ctx, attr.New("id", entity.Id))

	const sql = "INSERT INTO HELLO_DATA (ID, NAME, AGE) VALUES ($1, $2, $3)"
	var err error
	var params = []any{entity.Id, entity.Name, entity.Age}

	pgTx := r.getTransactionOrNil()
	if pgTx == nil {
		_, err = r.pool.Exec(ctx, sql, params...)
	} else {
		_, err = pgTx.Exec(ctx, sql, params...)
	}

	if err != nil {
		trace.InjectError(ctx, err)
		return nil, err
	}
	return entity, nil
}

func (r HelloRepository) FindById(ctx context.Context, id any) (*models.HelloData, error) {
	ctx, end := trace.Trace(ctx, trace.NameConfig(traceName, "FindById"))
	defer end()

	strId := id.(string)
	trace.InjectAttributes(ctx, attr.New("id", strId))
	const sql = "SELECT ID, NAME, AGE FROM HELLO_DATA WHERE ID = $1"
	pgTx := r.getTransactionOrNil()

	if pgTx == nil {
		row := r.pool.QueryRow(ctx, sql, strId)
		return r.mapRow(&row)
	}

	row := pgTx.QueryRow(ctx, sql, strId)
	return r.mapRow(&row)
}

func (r HelloRepository) ListAll(ctx context.Context) ([]models.HelloData, error) {
	ctx, end := trace.Trace(ctx, trace.NameConfig(traceName, "ListAll"))
	defer end()

	const sql = "SELECT ID, NAME, AGE FROM HELLO_DATA"
	var rows pgx.Rows
	var err error

	pgTx := r.getTransactionOrNil()
	if pgTx == nil {
		rows, err = r.pool.Query(ctx, sql)
	} else {
		rows, err = pgTx.Query(ctx, sql)
	}

	if err != nil {
		trace.InjectError(ctx, err)
		return nil, fmt.Errorf("list all hello data: %w", err)
	}
	defer rows.Close()

	result := make([]models.HelloData, 0)
	for rows.Next() {
		data, err := r.mapRowsNextValue(&rows)
		if err != nil {
			trace.InjectError(ctx, err)
			return nil, fmt.Errorf("scan hello data row: %w", err)
		}
		result = append(result, *data)
	}

	if err = rows.Err(); err != nil {
		trace.InjectError(ctx, err)
		return nil, fmt.Errorf("iterate hello data rows: %w", err)
	}

	return result, nil
}

func (r HelloRepository) getTransactionOrNil() pgx.Tx {
	if r.txRef == nil {
		return nil
	}
	return r.txRef.current()
}

func (r HelloRepository) mapRow(row *pgx.Row) (*models.HelloData, error) {
	result := new(models.HelloData)
	err := (*row).Scan(&result.Id, &result.Name, &result.Age)
	return result, err
}

func (r HelloRepository) mapRowsNextValue(rows *pgx.Rows) (*models.HelloData, error) {
	result := new(models.HelloData)
	err := (*rows).Scan(&result.Id, &result.Name, &result.Age)
	return result, err
}
