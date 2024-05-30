package hello

import (
	"context"
	"fmt"
	"main/internal/infra/observability/trace"
)

var tracer = trace.GetTracer("HelloService")

type Repository interface {
	Save(ctx context.Context, entity *HelloData) (*HelloData, error)
	FindById(ctx context.Context, id any) (*HelloData, error)
	ListAll(ctx context.Context) []HelloData
	BeginTransactionWithContext(ctx context.Context) (context.Context, error)
	Rollback(ctx context.Context)
	Commit(ctx context.Context)
	RunWithTransaction(ctx context.Context, callback func(context.Context) (*HelloData, error)) (*HelloData, error)
}

type HelloService struct {
	repository Repository
}

func NewService(repo Repository) HelloService {
	return HelloService{repository: repo}
}

func (s HelloService) Hello(ctx context.Context, id string, age int) string {
	ctx, spanOut := tracer.StartSpan(ctx, "Hello")
	spanOut.SetAttributes(trace.Attribute("id", id))
	defer spanOut.End()

	data, err := s.repository.RunWithTransaction(ctx, func(ctxTx context.Context) (*HelloData, error) {
		ctxTx, span := tracer.StartSpan(ctxTx, "HelloInTransaction")
		span.SetAttributes(trace.Attribute("id", id))
		defer span.End()
		newHello := HelloData{Id: id, Name: fmt.Sprintf("Bruno %v", id), Age: age}
		helloAdded, err := s.repository.Save(ctxTx, &newHello)
		if err != nil {
			span.SetError(err)
			return nil, err
		}
		helloFound, err := s.repository.FindById(ctxTx, helloAdded.Id)
		if err != nil {
			span.SetError(err)
			return nil, err
		}
		return helloFound, nil
	})
	if err != nil {
		spanOut.SetError(err)
		return err.Error()
	}
	return data.ToString()
}

// func (s HelloService) listAll(ctx context.Context) {
// 	list := s.repository.ListAll(ctx)
// 	fmt.Printf("Printing %v rows\n", len(list))
// 	for i := range list {
// 		fmt.Printf("Data: %v\n", list[i])
// 	}
// }
