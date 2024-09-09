package unitofwork

import (
	"context"

	"github.com/bruno303/study-topics/go-study/internal/hello/hellomodel"
	"github.com/bruno303/study-topics/go-study/internal/hellofile"
)

type Transaction interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
	HelloRepository() hellomodel.HelloRepository
	HelloFileService() *hellofile.HelloFileService
}

type UnitOfWork interface {
	BeginTransaction(context.Context) (Transaction, error)
}
