package hello

import (
	"context"
	"errors"
	"testing"
)

var (
	repo      fakeRepo                = fakeRepo{}
	txManager *fakeTransactionManager = &fakeTransactionManager{}
)

func TestHello(t *testing.T) {
	expected := HelloData{Id: "id", Name: "name", Age: 30}
	subject := NewService(txManager, repo)

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
	errorStr := "error xpto"
	txManager.response = func(ctx context.Context) (any, error) {
		return nil, errors.New(errorStr)
	}
	subject := NewService(txManager, repo)

	_, err := subject.Hello(context.Background(), "id", 18)
	if err == nil {
		t.Errorf("Expected error didn't occur: %v", err)
		return
	}
	if err.Error() != errorStr {
		t.Errorf("Error should be \n%s \nbut got \n%s", errorStr, err.Error())
	}
}

type fakeRepo struct{}
type fakeTransactionManager struct {
	response func(context.Context) (any, error)
}

func (r fakeRepo) Save(ctx context.Context, entity *HelloData) (*HelloData, error) {
	return entity, nil
}

func (r fakeRepo) FindById(ctx context.Context, id any) (*HelloData, error) {
	return &HelloData{Id: id.(string), Name: "name", Age: 30}, nil
}

func (r fakeRepo) ListAll(ctx context.Context) []HelloData {
	return make([]HelloData, 0)
}

func (r *fakeTransactionManager) Execute(ctx context.Context, callback func(context.Context) (any, error)) (any, error) {
	if r.response != nil {
		return r.response(ctx)
	}
	return callback(ctx)
}
