package hello

import (
	"context"
	"main/internal/infra"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PgxRepository struct {
	pool *pgxpool.Pool
	Repository
}

type transactionKey struct {
	name string
}

var txKey = transactionKey{name: "db-transaction"}

// Here we can replace the repository implementation just creating the one desired
//
// func NewRepository(ctx context.Context, container infra.Container) Repository {
// 	type Rp struct {
// 		Repository
// 	}
// 	return Rp{
// 		memdb.NewMemDbRepository[HelloData](),
// 	}
// }

func NewRepository(ctx context.Context, container infra.Container) Repository {
	return PgxRepository{
		pool: container.Pgxpool,
	}
}

func (r PgxRepository) Save(ctx context.Context, entity *HelloData) (*HelloData, error) {
	ent := *entity
	tx := r.getTransactionOrNil(ctx)
	var err error = nil
	if tx == nil {
		_, err = r.pool.Exec(ctx, "INSERT INTO HELLO_DATA (ID, NAME, AGE) VALUES ($1, $2, $3)", ent.Id, ent.Name, ent.Age)
	} else {
		_, err = (*tx).Exec(ctx, "INSERT INTO HELLO_DATA (ID, NAME, AGE) VALUES ($1, $2, $3)", ent.Id, ent.Name, ent.Age)
	}

	if err != nil {
		return nil, err
	}
	return entity, nil
}

func (r PgxRepository) FindById(ctx context.Context, id any) (*HelloData, error) {
	strId := id.(string)
	tx := r.getTransactionOrNil(ctx)

	if tx == nil {
		row := r.pool.QueryRow(ctx, "SELECT ID, NAME, AGE FROM HELLO_DATA WHERE ID = $1", strId)
		return r.mapRow(row)
	} else {
		row := (*tx).QueryRow(ctx, "SELECT ID, NAME, AGE FROM HELLO_DATA WHERE ID = $1", strId)
		return r.mapRow(row)
	}
}

func (r PgxRepository) ListAll(ctx context.Context) []HelloData {
	var rows pgx.Rows = nil
	var err error = nil
	tx := r.getTransactionOrNil(ctx)
	if tx == nil {
		rows, err = r.pool.Query(ctx, "SELECT ID, NAME, AGE FROM HELLO_DATA")
	} else {
		rows, err = (*tx).Query(ctx, "SELECT ID, NAME, AGE FROM HELLO_DATA")
	}

	if err != nil {
		return nil
	}

	result := make([]HelloData, 0)

	for rows.Next() {
		data, err := r.mapRow(rows)
		if err != nil {
			return make([]HelloData, 0)
		}
		result = append(result, *data)
	}
	return result
}

func (r PgxRepository) BeginTransactionWithContext(ctx context.Context) (context.Context, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	return context.WithValue(ctx, txKey, &tx), nil
}

func (r PgxRepository) Rollback(ctx context.Context) {
	tx := r.getTransactionOrNil(ctx)
	if tx != nil {
		(*tx).Rollback(ctx)
	}
}

func (r PgxRepository) Commit(ctx context.Context) {
	tx := r.getTransactionOrNil(ctx)
	if tx != nil {
		(*tx).Commit(ctx)
	}
}

func (r PgxRepository) getTransactionOrNil(ctx context.Context) *pgx.Tx {
	tx := ctx.Value(txKey)
	if tx == nil {
		return nil
	}
	return tx.(*pgx.Tx)
}

func (r PgxRepository) RunWithTransaction(ctx context.Context, callback func(context.Context) (*HelloData, error)) (*HelloData, error) {
	ctxTx, err := r.BeginTransactionWithContext(ctx)
	if err != nil {
		return nil, err
	}
	tx := r.getTransactionOrNil(ctxTx)
	result, err := callback(ctxTx)
	if err != nil {
		(*tx).Rollback(ctxTx)
	} else {
		(*tx).Commit(ctxTx)
	}
	return result, err
}

func (r PgxRepository) mapRow(row pgx.Row) (*HelloData, error) {
	result := new(HelloData)
	id := new(string)
	name := new(string)
	age := new(int)

	err := row.Scan(id, name, age)
	if err != nil {
		return nil, err
	}

	result.Id = *id
	result.Name = *name
	result.Age = *age
	return result, nil
}
