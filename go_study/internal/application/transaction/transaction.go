package transaction

import (
	"context"
)

//go:generate go tool mockgen -source=transaction.go -destination=mocks.go -package transaction

type (
	Transaction        any
	TransactionalFunc  func(ctx context.Context, tx Transaction) (any, error)
	TransactionManager interface {
		Execute(context.Context, Opts, TransactionalFunc) (any, error)
	}
	Opts struct {
		Transaction Transaction
		RequiresNew bool
	}
)

var EmptyOpts = Opts{
	Transaction: nil,
	RequiresNew: true,
}
