package hello

import (
	"context"
	"testing"
)

func TestHello(t *testing.T) {
	expected := "HelloData[id=id, name=name, age=30]"
	subject := NewService(&fakeRepo{})

	result := subject.Hello(context.Background(), "id", 18)
	if result != expected {
		t.Errorf("Result should be \n%s \nbut got \n%s", expected, result)
	}

}

type fakeRepo struct{}

func (r *fakeRepo) Save(ctx context.Context, entity *HelloData) (*HelloData, error) {
	return entity, nil
}

func (r *fakeRepo) FindById(ctx context.Context, id any) (*HelloData, error) {
	return &HelloData{Id: id.(string), Name: "name", Age: 30}, nil
}

func (r *fakeRepo) ListAll(ctx context.Context) []HelloData {
	return make([]HelloData, 0)
}

func (r *fakeRepo) BeginTransactionWithContext(ctx context.Context) (context.Context, error) {
	return ctx, nil
}

func (r *fakeRepo) Rollback(ctx context.Context) {}

func (r *fakeRepo) Commit(ctx context.Context) {}

func (r *fakeRepo) RunWithTransaction(ctx context.Context, callback func(context.Context) (*HelloData, error)) (*HelloData, error) {
	return callback(ctx)
}
