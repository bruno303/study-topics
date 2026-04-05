package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/bruno303/study-topics/go-study/internal/application/transaction"
	"github.com/bruno303/study-topics/go-study/internal/crosscutting/observability/trace"
)

type UnitOfWorkFactory interface {
	Create() transaction.UnitOfWork
}

type TransactionManager struct {
	uowFactory UnitOfWorkFactory
}

var _ transaction.TransactionManager = (*TransactionManager)(nil)

func NewTransactionManager(uowFactory UnitOfWorkFactory) TransactionManager {
	return TransactionManager{uowFactory: uowFactory}
}

func (f TransactionManager) WithinTx(ctx context.Context, fn transaction.TransactionCallback) error {
	ctx, end := trace.Trace(ctx, trace.NameConfig("TransactionManager", "WithinTx"))
	defer end()

	uow := f.uowFactory.Create()
	err := uow.Begin(ctx)
	if err != nil {
		trace.InjectError(ctx, err)
		return err
	}

	return f.execute(ctx, uow, fn)
}

func (f TransactionManager) execute(ctx context.Context, uow transaction.UnitOfWork, fn transaction.TransactionCallback) error {
	var callbackPanic any
	var callbackErr error
	func() {
		defer func() {
			if recovered := recover(); recovered != nil {
				callbackPanic = recovered
			}
		}()

		callbackErr = fn(ctx, uow)
	}()

	if callbackPanic != nil {
		f.rollbackBestEffort(ctx, uow)
		panic(callbackPanic)
	}

	if callbackErr != nil {
		rollbackErr := f.rollback(ctx, uow)
		if rollbackErr != nil {
			joinedErr := combineCallbackAndRollbackErr(callbackErr, rollbackErr)
			trace.InjectError(ctx, joinedErr)
			return joinedErr
		}
		return callbackErr
	}

	commitErr := f.commit(ctx, uow)
	if commitErr != nil {
		trace.InjectError(ctx, commitErr)
		return commitErr
	}

	return nil
}

func combineCallbackAndRollbackErr(callbackErr, rollbackErr error) error {
	if rollbackErr == nil {
		return callbackErr
	}
	return errors.Join(callbackErr, fmt.Errorf("rollback transaction: %w", rollbackErr))
}

func (f TransactionManager) commit(ctx context.Context, uow transaction.UnitOfWork) error {
	ctx, end := trace.Trace(ctx, trace.NameConfig("TransactionManager", "Commit"))
	defer end()

	err := uow.Commit(ctx)
	if err != nil {
		trace.InjectError(ctx, err)
		return err
	}

	return nil
}

func (f TransactionManager) rollback(ctx context.Context, uow transaction.UnitOfWork) error {
	ctx, end := trace.Trace(ctx, trace.NameConfig("TransactionManager", "Rollback"))
	defer end()

	err := uow.Rollback(ctx)
	if err != nil {
		trace.InjectError(ctx, err)
		return err
	}
	return nil
}

func (f TransactionManager) rollbackBestEffort(ctx context.Context, uow transaction.UnitOfWork) {
	defer func() {
		if recovered := recover(); recovered != nil {
			trace.InjectError(ctx, fmt.Errorf("panic during rollback: %v", recovered))
		}
	}()

	rollbackErr := f.rollback(ctx, uow)
	if rollbackErr != nil {
		trace.InjectError(ctx, rollbackErr)
	}
}
