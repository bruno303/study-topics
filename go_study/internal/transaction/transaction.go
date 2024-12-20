package transaction

import (
	"context"
)

type Transaction interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type TransactionManager[T any] interface {
	Execute(context.Context, func(context.Context) (T, error)) (T, error)
}
