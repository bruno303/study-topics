package transaction

import (
	"context"

	"github.com/bruno303/study-topics/go-study/internal/application/repository"
)

//go:generate go tool mockgen -source=transaction.go -destination=mocks.go -package transaction

type (
	UnitOfWork interface {
		HelloRepository() repository.HelloRepository
		OutboxRepository() repository.OutboxRepository
	}

	TransactionCallback func(context.Context, UnitOfWork) error

	TransactionManager interface {
		WithinTx(context.Context, TransactionOpts, TransactionCallback) error
	}

	TransactionOpts struct {
		Parent UnitOfWork
	}
)

func EmptyOpts() TransactionOpts {
	return TransactionOpts{}
}

func WithParent(parent UnitOfWork) TransactionOpts {
	return TransactionOpts{Parent: parent}
}
