package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/bruno303/study-topics/go-study/internal/application/transaction"
	"go.uber.org/mock/gomock"
)

func TestTransactionManager_WithinTx_WhenBeginFails_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedErr := errors.New("begin failed")
	uow := transaction.NewMockUnitOfWork(ctrl)
	uow.EXPECT().Begin(gomock.Any()).Return(expectedErr)

	uowFactory := NewMockUnitOfWorkFactory(ctrl)
	uowFactory.EXPECT().Create().Return(uow)
	manager := NewTransactionManager(uowFactory)

	callbackCalled := false
	err := manager.WithinTx(context.Background(), func(_ context.Context, _ transaction.UnitOfWork) error {
		callbackCalled = true
		return nil
	})

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected begin error %v, got %v", expectedErr, err)
	}
	if callbackCalled {
		t.Fatal("expected callback not to be called when begin fails")
	}
}

func TestTransactionManager_WithinTx_WhenCallbackSucceeds_Commits(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uow := transaction.NewMockUnitOfWork(ctrl)
	gomock.InOrder(
		uow.EXPECT().Begin(gomock.Any()).Return(nil),
		uow.EXPECT().Commit(gomock.Any()).Return(nil),
	)

	uowFactory := NewMockUnitOfWorkFactory(ctrl)
	uowFactory.EXPECT().Create().Return(uow)
	manager := NewTransactionManager(uowFactory)

	callbackCalled := false
	err := manager.WithinTx(context.Background(), func(_ context.Context, gotUow transaction.UnitOfWork) error {
		callbackCalled = true
		if gotUow != uow {
			t.Fatal("expected callback to receive factory-created unit of work")
		}
		return nil
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !callbackCalled {
		t.Fatal("expected callback to be called")
	}
}

func TestTransactionManager_WithinTx_WhenCallbackFails_RollsBackAndReturnsCallbackError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedErr := errors.New("callback failed")
	uow := transaction.NewMockUnitOfWork(ctrl)
	gomock.InOrder(
		uow.EXPECT().Begin(gomock.Any()).Return(nil),
		uow.EXPECT().Rollback(gomock.Any()).Return(nil),
	)

	uowFactory := NewMockUnitOfWorkFactory(ctrl)
	uowFactory.EXPECT().Create().Return(uow)
	manager := NewTransactionManager(uowFactory)

	err := manager.WithinTx(context.Background(), func(_ context.Context, _ transaction.UnitOfWork) error {
		return expectedErr
	})

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected callback error %v, got %v", expectedErr, err)
	}
}

func TestTransactionManager_WithinTx_WhenCallbackAndRollbackFail_ReturnsJoinedError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	callbackErr := errors.New("callback failed")
	rollbackErr := errors.New("rollback failed")
	uow := transaction.NewMockUnitOfWork(ctrl)
	gomock.InOrder(
		uow.EXPECT().Begin(gomock.Any()).Return(nil),
		uow.EXPECT().Rollback(gomock.Any()).Return(rollbackErr),
	)

	uowFactory := NewMockUnitOfWorkFactory(ctrl)
	uowFactory.EXPECT().Create().Return(uow)
	manager := NewTransactionManager(uowFactory)

	err := manager.WithinTx(context.Background(), func(_ context.Context, _ transaction.UnitOfWork) error {
		return callbackErr
	})

	if !errors.Is(err, callbackErr) {
		t.Fatalf("expected joined error to contain callback error, got %v", err)
	}
	if !errors.Is(err, rollbackErr) {
		t.Fatalf("expected joined error to contain rollback error, got %v", err)
	}
}

func TestTransactionManager_WithinTx_WhenCommitFails_ReturnsCommitError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedErr := errors.New("commit failed")
	uow := transaction.NewMockUnitOfWork(ctrl)
	gomock.InOrder(
		uow.EXPECT().Begin(gomock.Any()).Return(nil),
		uow.EXPECT().Commit(gomock.Any()).Return(expectedErr),
	)

	uowFactory := NewMockUnitOfWorkFactory(ctrl)
	uowFactory.EXPECT().Create().Return(uow)
	manager := NewTransactionManager(uowFactory)

	err := manager.WithinTx(context.Background(), func(_ context.Context, _ transaction.UnitOfWork) error {
		return nil
	})

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected commit error %v, got %v", expectedErr, err)
	}
}

func TestTransactionManager_WithinTx_WhenCallbackPanics_RollsBackAndRePanics(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uow := transaction.NewMockUnitOfWork(ctrl)
	gomock.InOrder(
		uow.EXPECT().Begin(gomock.Any()).Return(nil),
		uow.EXPECT().Rollback(gomock.Any()).Return(nil),
	)

	uowFactory := NewMockUnitOfWorkFactory(ctrl)
	uowFactory.EXPECT().Create().Return(uow)
	manager := NewTransactionManager(uowFactory)
	expectedPanic := "boom"

	var recovered any
	func() {
		defer func() {
			recovered = recover()
		}()

		_ = manager.WithinTx(context.Background(), func(_ context.Context, _ transaction.UnitOfWork) error {
			panic(expectedPanic)
		})
	}()

	if recovered != expectedPanic {
		t.Fatalf("expected panic %q, got %#v", expectedPanic, recovered)
	}
}

func TestCombineCallbackAndRollbackErr_WhenRollbackNil_ReturnsCallbackError(t *testing.T) {
	callbackErr := errors.New("callback failed")

	got := combineCallbackAndRollbackErr(callbackErr, nil)

	if !errors.Is(got, callbackErr) {
		t.Fatalf("expected callback error to be preserved, got %v", got)
	}
}

func TestCombineCallbackAndRollbackErr_WhenRollbackFails_JoinsErrors(t *testing.T) {
	callbackErr := errors.New("callback failed")
	rollbackErr := errors.New("rollback failed")

	got := combineCallbackAndRollbackErr(callbackErr, rollbackErr)

	if !errors.Is(got, callbackErr) {
		t.Fatalf("expected joined error to contain callback error, got %v", got)
	}
	if !errors.Is(got, rollbackErr) {
		t.Fatalf("expected joined error to contain rollback error, got %v", got)
	}
}
