package repository

import (
	"context"

	"github.com/bruno303/study-topics/go-study/internal/application/hello/models"
	applicationRepository "github.com/bruno303/study-topics/go-study/internal/application/repository"
	"github.com/bruno303/study-topics/go-study/internal/application/transaction"
	"github.com/bruno303/study-topics/go-study/internal/infra/database"
)

type memDbStorage struct {
	hello *database.MemDbRepository[models.HelloData]
}

type MemDbUnitOfWork struct {
	helloRepository applicationRepository.HelloRepository
}

var _ transaction.UnitOfWork = (*MemDbUnitOfWork)(nil)

func newMemDbStorage() *memDbStorage {
	return &memDbStorage{hello: database.NewMemDbRepository[models.HelloData]()}
}

func NewMemDbStorage() *memDbStorage {
	return newMemDbStorage()
}

func NewMemDbUnitOfWork(storage *memDbStorage) *MemDbUnitOfWork {
	if storage == nil {
		storage = newMemDbStorage()
	}

	return &MemDbUnitOfWork{
		helloRepository: NewHelloMemDbRepository(storage.hello),
	}
}

func (uow *MemDbUnitOfWork) Begin(context.Context) error {
	return nil
}

func (uow *MemDbUnitOfWork) Commit(context.Context) error {
	return nil
}

func (uow *MemDbUnitOfWork) Rollback(context.Context) error {
	return nil
}

func (uow *MemDbUnitOfWork) HelloRepository() applicationRepository.HelloRepository {
	return uow.helloRepository
}
