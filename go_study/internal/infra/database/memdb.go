package database

import (
	"context"
	"fmt"
	"main/internal/infra/observability/trace"
	"sync"
)

type transactionKeyMemDb struct{ key string }

type transaction struct {
	tx       any
	commited bool
}

var (
	txKeyMemDb  = transactionKeyMemDb{key: "my-app.repository.tx"}
	memdbTracer = trace.GetTracer("MemDbRepository")
)

type Keyer interface {
	Key() string
}

type Entity interface {
	Keyer
}

type MemDbRepository[E Entity] struct {
	data *sync.Map
}

func NewMemDbRepository[E Entity]() *MemDbRepository[E] {
	return &MemDbRepository[E]{
		data: &sync.Map{},
	}
}

func (repo MemDbRepository[E]) FindById(ctx context.Context, id any) (*E, error) {
	_, span := memdbTracer.StartSpan(ctx, "FindById")
	span.SetAttributes(trace.Attribute("id", id.(string)))
	defer span.End()
	value, ok := repo.data.Load(id)
	if !ok {
		err := fmt.Errorf("entity with id %v not found", id)
		span.SetError(err)
		return new(E), err
	}
	valuePtr := value.(E)
	return &valuePtr, nil
}

func (repo MemDbRepository[E]) Save(ctx context.Context, entity *E) (*E, error) {
	_, span := memdbTracer.StartSpan(ctx, "FindById")
	key := (*entity).Key()
	span.SetAttributes(trace.Attribute("id", key))
	defer span.End()
	repo.data.Store(key, entity)
	return entity, nil
}

func (repo MemDbRepository[E]) ListAll(ctx context.Context) []E {
	_, span := memdbTracer.StartSpan(ctx, "ListAll")
	defer span.End()
	list := []E{}
	repo.data.Range(func(key any, value any) bool {
		list = append(list, value.(E))
		return true
	})
	return list
}

func (repo MemDbRepository[E]) BeginTransactionWithContext(ctx context.Context) (context.Context, error) {
	ctx, span := memdbTracer.StartSpan(ctx, "BeginTransactionWithContext")
	defer span.End()
	return context.WithValue(ctx, txKeyMemDb, &transaction{tx: struct{}{}, commited: false}), nil
}

func (repo MemDbRepository[E]) Commit(ctx context.Context) {
	ctx, span := memdbTracer.StartSpan(ctx, "Commit")
	defer span.End()
	tx, ok := repo.getTransactionFromContext(ctx)
	if !ok {
		return
	}
	fmt.Println("Commiting")
	if tx.commited {
		return
	}
	tx.commited = true
	// commit logic here
}

func (repo MemDbRepository[E]) Rollback(ctx context.Context) {
	ctx, span := memdbTracer.StartSpan(ctx, "Rollback")
	defer span.End()
	tx, ok := repo.getTransactionFromContext(ctx)
	if !ok {
		return
	}
	fmt.Printf("Rolling back. Commited: %v\n", tx.commited)
	if tx.commited {
		return
	}
	// rollback logic here
}

func (repo MemDbRepository[E]) getTransactionFromContext(ctx context.Context) (*transaction, bool) {
	ctx, span := memdbTracer.StartSpan(ctx, "getTransactionFromContext")
	defer span.End()
	txValue := ctx.Value(txKeyMemDb)
	if txValue == nil {
		return new(transaction), false
	}
	return txValue.(*transaction), true
}

func (repo MemDbRepository[E]) RunWithTransaction(ctx context.Context, callback func(context.Context) (*E, error)) (*E, error) {
	ctx, span := memdbTracer.StartSpan(ctx, "RunWithTransaction")
	defer span.End()
	ctx, _ = repo.BeginTransactionWithContext(ctx)
	defer repo.Rollback(ctx)
	result, err := callback(ctx)
	if err != nil {
		span.SetError(err)
		return nil, err
	}
	repo.Commit(ctx)
	return result, nil
}
