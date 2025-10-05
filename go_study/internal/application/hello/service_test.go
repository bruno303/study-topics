package hello

import (
	"context"
	"errors"
	"testing"

	"github.com/bruno303/study-topics/go-study/internal/application/hello/models"
	"go.uber.org/mock/gomock"
)

var (
	ctrl    *gomock.Controller
	repo    *MockHelloRepository
	subject HelloService
)

func beforeEach(t *testing.T) {
	ctrl = gomock.NewController(t)
	repo = NewMockHelloRepository(ctrl)
	subject = NewService(repo)
}

func TestHello(t *testing.T) {
	beforeEach(t)
	expected := models.HelloData{Id: "id", Name: "Bruno id", Age: 18}

	repo.
		EXPECT().
		Save(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, entity *models.HelloData) (*models.HelloData, error) {
			return entity, nil
		}).Times(2)

	result, err := subject.Hello(context.Background(), HelloInput{Id: "id", Age: 18})
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

	repo.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil, errors.New(errorStr)).Times(1)

	_, err := subject.Hello(context.Background(), HelloInput{Id: "id", Age: 18})
	if err == nil {
		t.Errorf("Expected error didn't occur: %v", err)
		return
	}
	if err.Error() != errorStr {
		t.Errorf("Error should be \n%s \nbut got \n%s", errorStr, err.Error())
	}
}
