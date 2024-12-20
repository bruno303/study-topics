package hello

import (
	"context"
	"errors"
	"fmt"

	"github.com/bruno303/study-topics/go-study/internal/crosscutting/observability/trace"
	"github.com/bruno303/study-topics/go-study/internal/transaction"
)

type HelloRepository interface {
	Save(ctx context.Context, entity *HelloData) (*HelloData, error)
	FindById(ctx context.Context, id any) (*HelloData, error)
	ListAll(ctx context.Context) []HelloData
}

const traceName = "HelloService"

type HelloService struct {
	transactionManager transaction.TransactionManager[any]
	helloRepository    HelloRepository
}

func NewService(transactionManager transaction.TransactionManager[any], helloRepository HelloRepository) HelloService {
	return HelloService{transactionManager, helloRepository}
}

func (s HelloService) ListAll(ctx context.Context) ([]HelloData, error) {
	ctx, end := trace.Trace(ctx, trace.NameConfig(traceName, "ListAll"))
	defer end()
	return s.helloRepository.ListAll(ctx), nil
}

func (s HelloService) Hello(ctx context.Context, id string, age int) (HelloData, error) {
	var result HelloData
	ctx, end := trace.Trace(ctx, trace.NameConfig(traceName, "Hello"))
	defer end()

	data, err := s.transactionManager.Execute(ctx, func(ctxTx context.Context) (any, error) {

		ctxTx, end := trace.Trace(ctxTx, trace.NameConfig(traceName, "HelloTransactional"))
		defer end()

		newHello := HelloData{Id: id, Name: fmt.Sprintf("Bruno %v", id), Age: age}
		helloAdded, err := s.helloRepository.Save(ctxTx, &newHello)
		if err != nil {
			trace.InjectError(ctxTx, err)
			return nil, err
		}
		helloFound, err := s.helloRepository.FindById(ctxTx, helloAdded.Id)
		if err != nil {
			trace.InjectError(ctxTx, err)
			return nil, err
		}
		return helloFound, nil
	})
	if err != nil {
		trace.InjectError(ctx, err)
		return result, err
	}
	if value, ok := data.(*HelloData); ok {
		return *value, nil
	}
	return result, errors.New("invalid type returned")
}
