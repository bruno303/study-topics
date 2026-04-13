package repository

import (
	"context"

	"github.com/bruno303/study-topics/go-study/internal/application/hello/models"
	applicationRepository "github.com/bruno303/study-topics/go-study/internal/application/repository"
	"github.com/bruno303/study-topics/go-study/internal/application/transaction"
	"github.com/bruno303/study-topics/go-study/internal/crosscutting/observability/trace"
	"github.com/bruno303/study-topics/go-study/internal/infra/database"
)

type memDbStorage struct {
	hello  *database.MemDbRepository[models.HelloData]
	outbox *database.MemDbRepository[memDbOutboxRecord]
}

type memDbUnitOfWork struct {
	helloRepository  applicationRepository.HelloRepository
	outboxRepository applicationRepository.OutboxRepository
}

type MemDbTransactionManager struct {
	storage *memDbStorage
}

var _ transaction.TransactionManager = (*MemDbTransactionManager)(nil)
var _ transaction.UnitOfWork = (*memDbUnitOfWork)(nil)

func newMemDbStorage() *memDbStorage {
	return &memDbStorage{
		hello:  database.NewMemDbRepository[models.HelloData](),
		outbox: database.NewMemDbRepository[memDbOutboxRecord](),
	}
}

func NewMemDbStorage() *memDbStorage {
	return newMemDbStorage()
}

func NewMemDbTransactionManager(storage *memDbStorage) *MemDbTransactionManager {
	if storage == nil {
		storage = newMemDbStorage()
	}

	return &MemDbTransactionManager{storage: storage}
}

func (tm *MemDbTransactionManager) WithinTx(ctx context.Context, opts transaction.TransactionOpts, fn transaction.TransactionCallback) error {
	ctx, end := trace.Trace(ctx, trace.NameConfig("MemDbTransactionManager", "WithinTx"))
	defer end()

	if opts.Parent != nil {
		return fn(ctx, opts.Parent)
	}

	return fn(ctx, tm.newUnitOfWork())
}

func (tm *MemDbTransactionManager) newUnitOfWork() transaction.UnitOfWork {
	return &memDbUnitOfWork{
		helloRepository:  NewHelloMemDbRepository(tm.storage.hello),
		outboxRepository: NewOutboxMemDbRepository(tm.storage.outbox),
	}
}

func (uow *memDbUnitOfWork) HelloRepository() applicationRepository.HelloRepository {
	return uow.helloRepository
}

func (uow *memDbUnitOfWork) OutboxRepository() applicationRepository.OutboxRepository {
	return uow.outboxRepository
}
