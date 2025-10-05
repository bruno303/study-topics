package repository

import (
	"context"

	"github.com/bruno303/study-topics/go-study/internal/application/hello/models"
	"github.com/bruno303/study-topics/go-study/internal/crosscutting/observability/trace"
	"github.com/bruno303/study-topics/go-study/internal/crosscutting/observability/trace/attr"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OptimizedHelloRepository struct {
	pool *pgxpool.Pool
}

func NewOptimizedHelloRepository(pool *pgxpool.Pool) OptimizedHelloRepository {
	return OptimizedHelloRepository{
		pool: pool,
	}
}

func (r OptimizedHelloRepository) Save(ctx context.Context, entity *models.HelloData) (*models.HelloData, error) {
	ctx, end := trace.Trace(ctx, trace.NameConfig(traceName, "Save"))
	defer end()

	trace.InjectAttributes(ctx, attr.New("id", entity.Id))

	const sql = "INSERT INTO HELLO_DATA (ID, NAME, AGE) VALUES ($1, $2, $3)"
	var err error = nil
	var params = []any{entity.Id, entity.Name, entity.Age}

	tx := r.getTransactionOrNil(ctx)
	if tx == nil {
		_, err = r.pool.Exec(ctx, sql, params...)
		if err != nil {
			trace.InjectError(ctx, err)
			return nil, err
		}
	} else {
		tx.AddOperation(func(tx *pgx.Tx) error {
			_, e := (*tx).Exec(ctx, sql, params...)
			return e
		})
	}
	return entity, nil
}

func (r OptimizedHelloRepository) FindById(ctx context.Context, id any) (*models.HelloData, error) {
	ctx, end := trace.Trace(ctx, trace.NameConfig(traceName, "FindById"))
	defer end()

	strId := id.(string)
	trace.InjectAttributes(ctx, attr.New("id", strId))
	const sql = "SELECT ID, NAME, AGE FROM HELLO_DATA WHERE ID = $1"
	row := r.pool.QueryRow(ctx, sql, strId)
	return r.mapRow(&row)
}

func (r OptimizedHelloRepository) ListAll(ctx context.Context) []models.HelloData {
	ctx, end := trace.Trace(ctx, trace.NameConfig(traceName, "ListAll"))
	defer end()

	const sql = "SELECT ID, NAME, AGE FROM HELLO_DATA"
	var rows pgx.Rows = nil
	var err error = nil

	rows, err = r.pool.Query(ctx, sql)

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

func (r OptimizedHelloRepository) getTransactionOrNil(ctx context.Context) *OptimizedTransaction {
	value := ctx.Value(txCtxKey)
	if tx, ok := value.(*OptimizedTransaction); ok {
		return tx
	}
	return nil
}

func (r OptimizedHelloRepository) mapRow(row *pgx.Row) (*models.HelloData, error) {
	result := new(models.HelloData)
	err := (*row).Scan(&result.Id, &result.Name, &result.Age)
	return result, err
}

func (r OptimizedHelloRepository) mapRowsNextValue(rows *pgx.Rows) (*models.HelloData, error) {
	result := new(models.HelloData)
	err := (*rows).Scan(&result.Id, &result.Name, &result.Age)
	return result, err
}
