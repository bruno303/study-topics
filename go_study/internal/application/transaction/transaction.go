package transaction

import (
	"context"

	"github.com/bruno303/study-topics/go-study/internal/application/repository"
)

//go:generate go tool mockgen -source=transaction.go -destination=mocks.go -package transaction

type (
	UnitOfWork interface {
		Begin(context.Context) error
		Commit(context.Context) error
		Rollback(context.Context) error
		HelloRepository() repository.HelloRepository
	}

	TransactionCallback func(context.Context, UnitOfWork) error

	TransactionManager interface {
		WithinTx(context.Context, TransactionCallback) error
	}
)
