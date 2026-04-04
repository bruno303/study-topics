package transaction

import (
	"context"

	"github.com/bruno303/study-topics/go-study/internal/application/repository"
)

//go:generate go tool mockgen -source=transaction.go -destination=mocks.go -package transaction

type (
	Transaction any

	UnitOfWork interface {
		Begin(context.Context, Opts) error
		Commit(context.Context) error
		Rollback(context.Context) error
		HelloRepository() repository.HelloRepository
	}

	TransactionManager interface {
		WithinTx(context.Context, Opts, func(context.Context, UnitOfWork) error) error
	}

	Propagation uint8

	Opts struct {
		Propagation Propagation
	}
)

const (
	PropagationUnspecified Propagation = iota
	PropagationJoin
	PropagationRequiresNew
	PropagationNested
)

var EmptyOpts = RequiresNew()

func Join() Opts {
	return Opts{
		Propagation: PropagationJoin,
	}
}

func RequiresNew() Opts {
	return Opts{
		Propagation: PropagationRequiresNew,
	}
}

func Nested() Opts {
	return Opts{
		Propagation: PropagationNested,
	}
}

func (o Opts) EffectivePropagation() Propagation {
	if o.Propagation != PropagationUnspecified {
		return o.Propagation
	}
	return PropagationJoin
}
