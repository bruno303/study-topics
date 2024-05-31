package hello

import (
	"context"
	"fmt"
	"main/internal/crosscutting/observability"
)

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

	return observability.TraceWithResultAndAttr(
		ctx,
		"HelloService",
		"Hello",
		func(ctx context.Context, sm observability.SpanModifier) string {

			data, err := s.repository.RunWithTransaction(ctx, func(ctxTx context.Context) (*HelloData, error) {
				newHello := HelloData{Id: id, Name: fmt.Sprintf("Bruno %v", id), Age: age}
				helloAdded, err := s.repository.Save(ctxTx, &newHello)
				if err != nil {
					sm.HandleError(err)
					return nil, err
				}
				helloFound, err := s.repository.FindById(ctxTx, helloAdded.Id)
				if err != nil {
					sm.HandleError(err)
					return nil, err
				}
				return helloFound, nil
			})
			if err != nil {
				sm.HandleError(err)
				return err.Error()
			}
			return data.ToString()
		},
	)
}
