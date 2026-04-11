package hello

import (
	"context"
	"errors"
	"testing"

	"github.com/bruno303/study-topics/go-study/internal/application/hello/models"
	applicationRepository "github.com/bruno303/study-topics/go-study/internal/application/repository"
	"github.com/bruno303/study-topics/go-study/internal/application/transaction"
	"go.uber.org/mock/gomock"
)

type contextKey string

func TestHello_UsesUnitOfWorkAndReturnsFirstSavedEntity(t *testing.T) {
	ctrl := gomock.NewController(t)
	uow := transaction.NewMockUnitOfWork(ctrl)
	repos := transaction.NewMockRepositoryAccessor(ctrl)
	repo := applicationRepository.NewMockHelloRepository(ctrl)
	subject := NewService(uow)

	input := HelloInput{Id: "id", Age: 18}
	baseCtx := context.WithValue(t.Context(), contextKey("scope"), "outer")
	txCtx := context.WithValue(baseCtx, contextKey("scope"), "tx")

	savedEntities := make([]models.HelloData, 0, 2)

	uow.EXPECT().
		WithinTx(baseCtx, gomock.Any()).
		DoAndReturn(func(_ context.Context, fn transaction.TransactionCallback) error {
			return fn(txCtx, repos)
		})

	repos.EXPECT().HelloRepository().Return(repo).Times(2)
	repo.EXPECT().Save(txCtx, gomock.AssignableToTypeOf(&models.HelloData{})).DoAndReturn(
		func(ctx context.Context, entity *models.HelloData) (*models.HelloData, error) {
			if ctx != txCtx {
				t.Fatalf("expected save to use callback context")
			}
			savedEntities = append(savedEntities, *entity)
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
	if len(savedEntities) != 2 {
		t.Fatalf("expected two save calls, got %d", len(savedEntities))
	}
	if savedEntities[0] != expected {
		t.Fatalf("expected first saved entity %v, got %v", expected, savedEntities[0])
	}
}

func TestHello_WhenRepositoryReturnsError_PropagatesErrorFromWithinTxCallback(t *testing.T) {
	ctrl := gomock.NewController(t)
	uow := transaction.NewMockUnitOfWork(ctrl)
	repos := transaction.NewMockRepositoryAccessor(ctrl)
	repo := applicationRepository.NewMockHelloRepository(ctrl)
	subject := NewService(uow)

	expectedErr := errors.New("save failed")

	uow.EXPECT().
		WithinTx(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn transaction.TransactionCallback) error {
			return fn(ctx, repos)
		})

	repos.EXPECT().HelloRepository().Return(repo)
	repo.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil, expectedErr)

	_, err := subject.Hello(t.Context(), HelloInput{Id: "id", Age: 18})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
}

func TestHello_WhenWithinTxFails_PropagatesError(t *testing.T) {
	ctrl := gomock.NewController(t)
	uow := transaction.NewMockUnitOfWork(ctrl)
	subject := NewService(uow)

	expectedErr := errors.New("begin tx failed")
	uow.EXPECT().
		WithinTx(gomock.Any(), gomock.Any()).
		Return(expectedErr)

	_, err := subject.Hello(t.Context(), HelloInput{Id: "id", Age: 18})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
}

func TestListAll_UsesUnitOfWorkAndReturnsResult(t *testing.T) {
	ctrl := gomock.NewController(t)
	uow := transaction.NewMockUnitOfWork(ctrl)
	repos := transaction.NewMockRepositoryAccessor(ctrl)
	repo := applicationRepository.NewMockHelloRepository(ctrl)
	subject := NewService(uow)

	expected := []models.HelloData{{Id: "id-1", Name: "Bruno id-1", Age: 18}}
	baseCtx := context.WithValue(t.Context(), contextKey("scope"), "outer")
	txCtx := context.WithValue(baseCtx, contextKey("scope"), "tx")

	uow.EXPECT().
		WithinTx(baseCtx, gomock.Any()).
		DoAndReturn(func(_ context.Context, fn transaction.TransactionCallback) error {
			return fn(txCtx, repos)
		})

	repos.EXPECT().HelloRepository().Return(repo)
	repo.EXPECT().ListAll(txCtx).Return(expected, nil)

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
	uow := transaction.NewMockUnitOfWork(ctrl)
	repos := transaction.NewMockRepositoryAccessor(ctrl)
	repo := applicationRepository.NewMockHelloRepository(ctrl)
	subject := NewService(uow)

	expectedErr := errors.New("repository list error")

	uow.EXPECT().
		WithinTx(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn transaction.TransactionCallback) error {
			return fn(ctx, repos)
		})

	repos.EXPECT().HelloRepository().Return(repo)
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
	uow := transaction.NewMockUnitOfWork(ctrl)
	subject := NewService(uow)

	expectedErr := errors.New("tx manager failed")
	uow.EXPECT().
		WithinTx(gomock.Any(), gomock.Any()).
		Return(expectedErr)

	_, err := subject.ListAll(t.Context())
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
}
