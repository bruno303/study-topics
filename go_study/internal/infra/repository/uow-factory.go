package repository

import "github.com/bruno303/study-topics/go-study/internal/application/transaction"

type UnitOfWorkFactoryImpl struct {
	config *PgxUnitOfWorkConfig
}

type MemDbUnitOfWorkFactoryImpl struct {
	storage *memDbStorage
}

func NewUnitOfWorkFactory(config *PgxUnitOfWorkConfig) UnitOfWorkFactoryImpl {
	return UnitOfWorkFactoryImpl{config: config}
}

func NewMemDbUnitOfWorkFactory(storage ...*memDbStorage) MemDbUnitOfWorkFactoryImpl {
	sharedStorage := newMemDbStorage()
	if len(storage) > 0 && storage[0] != nil {
		sharedStorage = storage[0]
	}

	return MemDbUnitOfWorkFactoryImpl{storage: sharedStorage}
}

func (f UnitOfWorkFactoryImpl) Create() transaction.UnitOfWork {
	return NewPgxUnitOfWork(f.config)
}

func (f MemDbUnitOfWorkFactoryImpl) Create() transaction.UnitOfWork {
	return NewMemDbUnitOfWork(f.storage)
}
