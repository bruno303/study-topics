package memdb

import (
	"context"
	"fmt"
	"sync"
)

type transactionKey struct{ key string }

type transaction struct {
	tx       any
	commited bool
}

var txKey = transactionKey{key: "my-app.repository.tx"}

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
	repo.data.Store((*entity).Key(), entity)
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
	return context.WithValue(ctx, txKey, &transaction{tx: struct{}{}, commited: false}), nil
}

func (repo MemDbRepository[E]) Commit(ctx context.Context) {
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
	txValue := ctx.Value(txKey)
	if txValue == nil {
		return new(transaction), false
	}
	return txValue.(*transaction), true
}

func (repo MemDbRepository[E]) RunWithTransaction(ctx context.Context, callback func(context.Context) (*E, error)) (*E, error) {
	ctxTx, _ := repo.BeginTransactionWithContext(ctx)
	defer repo.Rollback(ctxTx)
	result, err := callback(ctxTx)
	if err != nil {
		return nil, err
	}
	repo.Commit(ctxTx)
	return result, nil
}
