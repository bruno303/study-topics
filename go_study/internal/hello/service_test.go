package hello

import (
	"context"
	"errors"
	"testing"

	"github.com/bruno303/study-topics/go-study/internal/transaction"
	"go.uber.org/mock/gomock"
)

var (
	ctrl      *gomock.Controller
	txManager *transaction.MockTransactionManager
	repo      *MockHelloRepository
	subject   HelloService
)

func beforeEach(t *testing.T) {
	ctrl = gomock.NewController(t)
	txManager = transaction.NewMockTransactionManager(ctrl)
	repo = NewMockHelloRepository(ctrl)
	subject = NewService(txManager, repo)
}

func TestHello(t *testing.T) {
	beforeEach(t)
	expected := HelloData{Id: "id", Name: "Bruno id", Age: 18}

	txManager.
		EXPECT().
		Execute(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, callback func(context.Context) (any, error)) (any, error) {
			return callback(ctx)
		}).Times(1)

	repo.
		EXPECT().
		Save(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, entity *HelloData) (*HelloData, error) {
			return entity, nil
		}).Times(2)

	result, err := subject.Hello(context.Background(), "id", 18)
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
		Execute(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, callback func(context.Context) (any, error)) (any, error) {
			return nil, errors.New(errorStr)
		}).
		Times(1)

	_, err := subject.Hello(context.Background(), "id", 18)
	if err == nil {
		t.Errorf("Expected error didn't occur: %v", err)
		return
	}
	if err.Error() != errorStr {
		t.Errorf("Error should be \n%s \nbut got \n%s", errorStr, err.Error())
	}
}
