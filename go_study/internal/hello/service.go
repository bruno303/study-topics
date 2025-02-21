package hello

import (
	"context"
	"errors"
	"fmt"

	"github.com/bruno303/study-topics/go-study/internal/crosscutting/observability/trace"
	"github.com/bruno303/study-topics/go-study/internal/transaction"
	"github.com/google/uuid"
)

type HelloRepository interface {
	Save(ctx context.Context, entity *HelloData) (*HelloData, error)
	FindById(ctx context.Context, id any) (*HelloData, error)
	ListAll(ctx context.Context) []HelloData
}

const traceName = "HelloService"

type HelloService struct {
	transactionManager transaction.TransactionManager
	helloRepository    HelloRepository
}

func NewService(
	transactionManager transaction.TransactionManager,
	helloRepository HelloRepository,
) HelloService {
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
		hello1 := HelloData{Id: id, Name: fmt.Sprintf("Bruno %v", id), Age: age}
		_, err := s.helloRepository.Save(ctxTx, &hello1)
		if err != nil {
			trace.InjectError(ctxTx, err)
			return nil, err
		}
		id2 := uuid.NewString()
		hello2 := HelloData{Id: id2, Name: fmt.Sprintf("Bruno %v", id2), Age: age}
		_, err = s.helloRepository.Save(ctxTx, &hello2)
		if err != nil {
			trace.InjectError(ctxTx, err)
			return nil, err
		}
		return &hello1, nil
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
