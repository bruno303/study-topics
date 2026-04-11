package transaction

import (
	"context"

	"github.com/bruno303/study-topics/go-study/internal/application/repository"
)

//go:generate go tool mockgen -source=transaction.go -destination=mocks.go -package transaction

type (
	RepositoryAccessor interface {
		HelloRepository() repository.HelloRepository
	}

	TransactionCallback func(context.Context, RepositoryAccessor) error

	UnitOfWork interface {
		WithinTx(context.Context, TransactionCallback) error
	}
)
