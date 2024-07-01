package hello

import (
	"context"
	"fmt"

	"github.com/bruno303/study-topics/go-study/internal/crosscutting/observability/trace"
)

const traceName = "HelloService"

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

func (s HelloService) Hello2(ctx context.Context, id string, age int) string {
	ctx, end := trace.Trace(ctx, trace.NameConfig(traceName, "Hello2"))
	defer end()

	ctx, err := s.repository.BeginTransactionWithContext(ctx)
	defer s.repository.Rollback(ctx)

	if err != nil {
		trace.InjectError(ctx, err)
		return err.Error()
	}

	newHello := HelloData{Id: id, Name: fmt.Sprintf("Bruno %v", id), Age: age}
	helloAdded, err := s.repository.Save(ctx, &newHello)
	if err != nil {
		trace.InjectError(ctx, err)
		return err.Error()
	}
	helloFound, err := s.repository.FindById(ctx, helloAdded.Id)
	if err != nil {
		trace.InjectError(ctx, err)
		return err.Error()
	}
	s.repository.Commit(ctx)
	return helloFound.ToString()
}

func (s HelloService) Hello(ctx context.Context, id string, age int) string {
	ctx, end := trace.Trace(ctx, trace.NameConfig(traceName, "Hello"))
	defer end()

	data, err := s.repository.RunWithTransaction(ctx, func(ctxTx context.Context) (*HelloData, error) {

		ctxTx, end := trace.Trace(ctxTx, trace.NameConfig(traceName, "HelloTransactional"))
		defer end()

		newHello := HelloData{Id: id, Name: fmt.Sprintf("Bruno %v", id), Age: age}
		helloAdded, err := s.repository.Save(ctxTx, &newHello)
		if err != nil {
			trace.InjectError(ctxTx, err)
			return nil, err
		}
		helloFound, err := s.repository.FindById(ctxTx, helloAdded.Id)
		if err != nil {
			trace.InjectError(ctxTx, err)
			return nil, err
		}
		return helloFound, nil
	})
	if err != nil {
		trace.InjectError(ctx, err)
		return err.Error()
	}
	return data.ToString()
}
