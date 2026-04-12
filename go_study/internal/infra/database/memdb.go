package database

import (
	"context"
	"fmt"
	"sync"
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
	repo.data.Store(key, *entity)
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
