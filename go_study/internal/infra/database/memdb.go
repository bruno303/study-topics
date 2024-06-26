package database

import (
	"context"
	"fmt"
	"sync"

	"github.com/bruno303/study-topics/go-study/internal/crosscutting/observability/log"
)

type transactionKeyMemDb struct{ key string }

type transaction struct {
	tx       any
	commited bool
}

var (
	txKeyMemDb = transactionKeyMemDb{key: "my-app.repository.tx"}
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
	value, ok := repo.data.Load(id)
	if !ok {
		return new(E), fmt.Errorf("entity with id %v not found", id)
	}
	valuePtr := value.(E)
	return &valuePtr, nil
}

func (repo MemDbRepository[E]) Save(ctx context.Context, entity *E) (*E, error) {
	key := (*entity).Key()
	repo.data.Store(key, entity)
	return entity, nil
}

func (repo MemDbRepository[E]) ListAll(ctx context.Context) []E {
	list := []E{}
	repo.data.Range(func(key any, value any) bool {
		list = append(list, value.(E))
		return true
	})
	return list
}

func (repo MemDbRepository[E]) BeginTransactionWithContext(ctx context.Context) (context.Context, error) {
	return context.WithValue(ctx, txKeyMemDb, &transaction{tx: struct{}{}, commited: false}), nil
}

func (repo MemDbRepository[E]) Commit(ctx context.Context) {
	tx, ok := repo.getTransactionFromContext(ctx)
	if !ok {
		return
	}
	log.Log().Info(ctx, "Commiting")
	if tx.commited {
		return
	}
	tx.commited = true
	// commit logic here
}

func (repo MemDbRepository[E]) Rollback(ctx context.Context) {
	tx, ok := repo.getTransactionFromContext(ctx)
	if !ok {
		return
	}
	log.Log().Info(ctx, "Rolling back. Commited: %v", tx.commited)
	if tx.commited {
		return
	}
	// rollback logic here
}

func (repo MemDbRepository[E]) getTransactionFromContext(ctx context.Context) (*transaction, bool) {
	txValue := ctx.Value(txKeyMemDb)
	if txValue == nil {
		return new(transaction), false
	}
	return txValue.(*transaction), true
}

func (repo MemDbRepository[E]) RunWithTransaction(ctx context.Context, callback func(context.Context) (*E, error)) (*E, error) {
	ctx, _ = repo.BeginTransactionWithContext(ctx)
	defer repo.Rollback(ctx)
	result, err := callback(ctx)
	if err != nil {
		return nil, err
	}
	repo.Commit(ctx)
	return result, nil
}
