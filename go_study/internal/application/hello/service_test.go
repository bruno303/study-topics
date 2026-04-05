package hello

import (
	"context"
	"errors"
	"testing"

	"github.com/bruno303/study-topics/go-study/internal/application/hello/models"
	"github.com/bruno303/study-topics/go-study/internal/application/repository"
	"github.com/bruno303/study-topics/go-study/internal/application/transaction"
	"go.uber.org/mock/gomock"
)

type contextKey string

func TestHello_UsesTransactionManagerAndReturnsFirstSavedEntity(t *testing.T) {
	ctrl := gomock.NewController(t)
	txManager := transaction.NewMockTransactionManager(ctrl)
	uow := transaction.NewMockUnitOfWork(ctrl)
	repo := repository.NewMockHelloRepository(ctrl)
	subject := NewService(txManager)

	input := HelloInput{Id: "id", Age: 18}
	baseCtx := context.WithValue(t.Context(), contextKey("scope"), "outer")
	txCtx := context.WithValue(baseCtx, contextKey("scope"), "tx")

	txManager.EXPECT().
		WithinTx(baseCtx, gomock.Any()).
		DoAndReturn(func(_ context.Context, fn func(context.Context, transaction.UnitOfWork) error) error {
			return fn(txCtx, uow)
		})

	uow.EXPECT().HelloRepository().Return(repo).Times(2)

	repo.EXPECT().Save(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, entity *models.HelloData) (*models.HelloData, error) {
			if ctx != txCtx {
				t.Fatalf("expected save to use callback context")
			}
			return entity, nil
		},
	).Times(2)

	result, err := subject.Hello(baseCtx, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := models.HelloData{Id: "id", Name: "Bruno id", Age: 18}
	if result != expected {
		t.Fatalf("expected %v, got %v", expected, result)
	}
}

func TestHello_WhenRepositoryReturnsError_PropagatesErrorFromWithinTxCallback(t *testing.T) {
	ctrl := gomock.NewController(t)
	txManager := transaction.NewMockTransactionManager(ctrl)
	uow := transaction.NewMockUnitOfWork(ctrl)
	repo := repository.NewMockHelloRepository(ctrl)
	subject := NewService(txManager)

	expectedErr := errors.New("save failed")

	txManager.EXPECT().
		WithinTx(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(context.Context, transaction.UnitOfWork) error) error {
			return fn(ctx, uow)
		})

	uow.EXPECT().HelloRepository().Return(repo)
	repo.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil, expectedErr)

	_, err := subject.Hello(t.Context(), HelloInput{Id: "id", Age: 18})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
}

func TestHello_WhenWithinTxFails_PropagatesError(t *testing.T) {
	ctrl := gomock.NewController(t)
	txManager := transaction.NewMockTransactionManager(ctrl)
	subject := NewService(txManager)

	expectedErr := errors.New("begin tx failed")
	txManager.EXPECT().
		WithinTx(gomock.Any(), gomock.Any()).
		Return(expectedErr)

	_, err := subject.Hello(t.Context(), HelloInput{Id: "id", Age: 18})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
}

func TestListAll_UsesTransactionManagerAndReturnsResult(t *testing.T) {
	ctrl := gomock.NewController(t)
	txManager := transaction.NewMockTransactionManager(ctrl)
	uow := transaction.NewMockUnitOfWork(ctrl)
	repo := repository.NewMockHelloRepository(ctrl)
	subject := NewService(txManager)

	expected := []models.HelloData{{Id: "id-1", Name: "Bruno id-1", Age: 18}}
	baseCtx := context.WithValue(t.Context(), contextKey("scope"), "outer")
	txCtx := context.WithValue(baseCtx, contextKey("scope"), "tx")
	innerTxCtx := context.WithValue(txCtx, contextKey("scope"), "inner")

	gomock.InOrder(
		txManager.EXPECT().
			WithinTx(baseCtx, gomock.Any()).
			DoAndReturn(func(_ context.Context, fn func(context.Context, transaction.UnitOfWork) error) error {
				return fn(txCtx, uow)
			}),
		txManager.EXPECT().
			WithinTx(txCtx, gomock.Any()).
			DoAndReturn(func(_ context.Context, fn func(context.Context, transaction.UnitOfWork) error) error {
				return fn(innerTxCtx, uow)
			}),
	)

	uow.EXPECT().HelloRepository().Return(repo).Times(2)
	callNumber := 0
	repo.EXPECT().ListAll(gomock.Any()).Times(2).DoAndReturn(func(ctx context.Context) ([]models.HelloData, error) {
		callNumber++
		switch callNumber {
		case 1:
			if ctx != txCtx {
				t.Fatalf("expected outer list to use callback context")
			}
		case 2:
			if ctx != innerTxCtx {
				t.Fatalf("expected inner list to use nested callback context")
			}
		default:
			t.Fatalf("unexpected call number %d", callNumber)
		}

		return expected, nil
	})

	result, err := subject.ListAll(baseCtx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != len(expected) || result[0] != expected[0] {
		t.Fatalf("expected %v, got %v", expected, result)
	}
}

func TestListAll_WhenRepositoryReturnsError_PropagatesErrorFromWithinTxCallback(t *testing.T) {
	ctrl := gomock.NewController(t)
	txManager := transaction.NewMockTransactionManager(ctrl)
	uow := transaction.NewMockUnitOfWork(ctrl)
	repo := repository.NewMockHelloRepository(ctrl)
	subject := NewService(txManager)

	expectedErr := errors.New("repository list error")

	txManager.EXPECT().
		WithinTx(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(context.Context, transaction.UnitOfWork) error) error {
			return fn(ctx, uow)
		})

	uow.EXPECT().HelloRepository().Return(repo)
	repo.EXPECT().ListAll(gomock.Any()).Return(nil, expectedErr)

	result, err := subject.ListAll(t.Context())
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
	if result != nil {
		t.Fatalf("expected nil result, got %v", result)
	}
}

func TestListAll_WhenWithinTxFails_PropagatesError(t *testing.T) {
	ctrl := gomock.NewController(t)
	txManager := transaction.NewMockTransactionManager(ctrl)
	subject := NewService(txManager)

	expectedErr := errors.New("tx manager failed")
	txManager.EXPECT().
		WithinTx(gomock.Any(), gomock.Any()).
		Return(expectedErr)

	_, err := subject.ListAll(t.Context())
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
}
