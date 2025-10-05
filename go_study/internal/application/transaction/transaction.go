package transaction

import (
	"context"
)

//go:generate mockgen -source=transaction.go -destination=mocks.go -package transaction

type Transaction interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type TransactionManager interface {
	Execute(context.Context, func(context.Context) (any, error)) (any, error)
}
