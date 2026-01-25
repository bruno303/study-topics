package hello

import (
	"context"
	"errors"
	"testing"

	"github.com/bruno303/study-topics/go-study/internal/application/hello/models"
	"github.com/bruno303/study-topics/go-study/internal/application/transaction"
	"go.uber.org/mock/gomock"
)

var (
	ctrl      *gomock.Controller
	repo      *MockHelloRepository
	txManager *transaction.MockTransactionManager
	tx        any
	subject   HelloService
	opts      transaction.Opts
)

func beforeEach(t *testing.T) {
	ctrl = gomock.NewController(t)
	repo = NewMockHelloRepository(ctrl)
	txManager = transaction.NewMockTransactionManager(ctrl)
	tx = struct{}{}
	subject = NewService(repo, txManager)
	opts = transaction.Opts{Transaction: nil, RequiresNew: true}
}

func TestHello(t *testing.T) {
	beforeEach(t)
	expected := models.HelloData{Id: "id", Name: "Bruno id", Age: 18}

	txManager.
		EXPECT().
		Execute(gomock.Any(), gomock.Eq(opts), gomock.Any()).
		DoAndReturn(func(ctx context.Context, opts transaction.Opts, fn transaction.TransactionalFunc) (any, error) {
			return fn(ctx, tx)
		}).Times(1)

	repo.
		EXPECT().
		Save(gomock.Any(), gomock.Any(), gomock.Eq(tx)).
		DoAndReturn(func(ctx context.Context, entity *models.HelloData, tx transaction.Transaction) (*models.HelloData, error) {
			return entity, nil
		}).Times(2)

	result, err := subject.Hello(t.Context(), HelloInput{Id: "id", Age: 18})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}
	if result != expected {
		t.Errorf("Result should be \n%v \nbut got \n%v", expected, result)
	}
}

func TestHelloWithError(t *testing.T) {
	beforeEach(t)
	errorStr := "error xpto"

	txManager.
		EXPECT().
		Execute(gomock.Any(), gomock.Eq(opts), gomock.Any()).
		DoAndReturn(func(ctx context.Context, opts transaction.Opts, fn transaction.TransactionalFunc) (any, error) {
			return fn(ctx, tx)
		}).Times(1)

	repo.EXPECT().Save(gomock.Any(), gomock.Any(), gomock.Eq(tx)).Return(nil, errors.New(errorStr)).Times(1)

	_, err := subject.Hello(t.Context(), HelloInput{Id: "id", Age: 18})
	if err == nil {
		t.Errorf("Expected error didn't occur: %v", err)
		return
	}
	if err.Error() != errorStr {
		t.Errorf("Error should be \n%s \nbut got \n%s", errorStr, err.Error())
	}
}
