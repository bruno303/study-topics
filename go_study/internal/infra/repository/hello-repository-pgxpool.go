package repository

import (
	"context"

	"github.com/bruno303/study-topics/go-study/internal/application/hello/models"
	"github.com/bruno303/study-topics/go-study/internal/application/transaction"
	"github.com/bruno303/study-topics/go-study/internal/crosscutting/observability/trace"
	"github.com/bruno303/study-topics/go-study/internal/crosscutting/observability/trace/attr"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type HelloRepository struct {
	pool *pgxpool.Pool
}

const traceName = "HelloRepository"

func NewHelloPgxRepository(pool *pgxpool.Pool) HelloRepository {
	return HelloRepository{
		pool: pool,
	}
}

func (r HelloRepository) Save(ctx context.Context, entity *models.HelloData, tx transaction.Transaction) (*models.HelloData, error) {
	ctx, end := trace.Trace(ctx, trace.NameConfig(traceName, "Save"))
	defer end()

	trace.InjectAttributes(ctx, attr.New("id", entity.Id))

	const sql = "INSERT INTO HELLO_DATA (ID, NAME, AGE) VALUES ($1, $2, $3)"
	var err error
	var params = []any{entity.Id, entity.Name, entity.Age}

	pgTx := r.getTransactionOrNil(tx)
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

func (r HelloRepository) FindById(ctx context.Context, id any, tx transaction.Transaction) (*models.HelloData, error) {
	ctx, end := trace.Trace(ctx, trace.NameConfig(traceName, "FindById"))
	defer end()

	strId := id.(string)
	trace.InjectAttributes(ctx, attr.New("id", strId))
	const sql = "SELECT ID, NAME, AGE FROM HELLO_DATA WHERE ID = $1"
	pgTx := r.getTransactionOrNil(tx)

	if pgTx == nil {
		row := r.pool.QueryRow(ctx, sql, strId)
		return r.mapRow(&row)
	} else {
		row := pgTx.QueryRow(ctx, sql, strId)
		return r.mapRow(&row)
	}
}

func (r HelloRepository) ListAll(ctx context.Context, tx transaction.Transaction) []models.HelloData {
	ctx, end := trace.Trace(ctx, trace.NameConfig(traceName, "ListAll"))
	defer end()

	const sql = "SELECT ID, NAME, AGE FROM HELLO_DATA"
	var rows pgx.Rows = nil
	var err error

	pgTx := r.getTransactionOrNil(tx)
	if tx == nil {
		rows, err = r.pool.Query(ctx, sql)
	} else {
		rows, err = pgTx.Query(ctx, sql)
	}

	if err != nil {
		trace.InjectError(ctx, err)
		return nil
	}

	result := make([]models.HelloData, 0)

	for rows.Next() {
		data, err := r.mapRowsNextValue(&rows)
		if err != nil {
			return make([]models.HelloData, 0)
		}
		result = append(result, *data)
	}
	return result
}

func (r HelloRepository) getTransactionOrNil(tx transaction.Transaction) pgx.Tx {
	appTx, ok := tx.(pgx.Tx)
	if !ok {
		return nil
	}
	return appTx
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
