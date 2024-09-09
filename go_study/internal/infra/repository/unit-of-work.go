package repository

import (
	"context"

	"github.com/bruno303/study-topics/go-study/internal/crosscutting/observability/trace"
	"github.com/bruno303/study-topics/go-study/internal/hello/hellomodel"
	"github.com/bruno303/study-topics/go-study/internal/hellofile"
	"github.com/bruno303/study-topics/go-study/internal/unitofwork"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type (
	UnitOfWork struct {
		config *UnitOfWorkConfig
	}
	ApplicationTransaction struct {
		postgreTransaction *pgx.Tx
		helloRepository    hellomodel.HelloRepository
		helloFileService   hellofile.HelloFileService
	}
	UnitOfWorkConfig struct {
		Pool                   *pgxpool.Pool
		HelloRepositoryFactory func(ctx context.Context, tx *pgx.Tx) hellomodel.HelloRepository
	}
)

func NewUnitOfWork(cfg *UnitOfWorkConfig) UnitOfWork {
	return UnitOfWork{cfg}
}

func (u UnitOfWork) BeginTransaction(ctx context.Context) (unitofwork.Transaction, error) {
	ctx, end := trace.Trace(ctx, trace.NameConfig("UnitOfWork", "BeginTransaction"))
	defer end()

	pTx, err := u.config.Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	tx := &ApplicationTransaction{
		helloRepository:    u.config.HelloRepositoryFactory(ctx, &pTx),
		helloFileService:   hellofile.NewHelloFileService(),
		postgreTransaction: &pTx,
	}
	return tx, nil
}

func (u *ApplicationTransaction) HelloRepository() hellomodel.HelloRepository {
	return u.helloRepository
}

func (u *ApplicationTransaction) HelloFileService() *hellofile.HelloFileService {
	return &u.helloFileService
}

func (u *ApplicationTransaction) Commit(ctx context.Context) error {
	ctx, end := trace.Trace(ctx, trace.NameConfig("UnitOfWork", "Commit"))
	defer end()
	if u.postgreTransaction != nil {
		return (*u.postgreTransaction).Commit(ctx)
	}
	return nil
}

func (u *ApplicationTransaction) Rollback(ctx context.Context) error {
	ctx, end := trace.Trace(ctx, trace.NameConfig("UnitOfWork", "Rollback"))
	defer end()
	if u.postgreTransaction != nil {
		if err := (*u.postgreTransaction).Rollback(ctx); err != nil {
			return err
		}
	}
	if err := u.helloFileService.Rollback(ctx); err != nil {
		return err
	}
	return nil
}
