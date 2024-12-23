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

type FileRepository interface {
	WriteFile(ctx context.Context, path string, content []byte) error
}

const traceName = "HelloService"

type HelloService struct {
	transactionManager transaction.TransactionManager
	helloRepository    HelloRepository
	fileRepository     FileRepository
}

func NewService(
	transactionManager transaction.TransactionManager,
	helloRepository HelloRepository,
	fileRepository FileRepository,
) HelloService {
	return HelloService{transactionManager, helloRepository, fileRepository}
}

func (s HelloService) ListAll(ctx context.Context) ([]HelloData, error) {
	ctx, end := trace.Trace(ctx, trace.NameConfig(traceName, "ListAll"))
	defer end()
	return s.helloRepository.ListAll(ctx), nil
}

func (s HelloService) HelloWriting(ctx context.Context, path string, content string) error {
	ctx, end := trace.Trace(ctx, trace.NameConfig(traceName, "HelloWriting"))
	defer end()
	return s.fileRepository.WriteFile(ctx, path, []byte(content))
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
		// if err = s.fileRepository.WriteFile(ctxTx, fmt.Sprintf("files/%s.txt", helloFound.Id), []byte(helloFound.ToString())); err != nil {
		// 	return nil, err
		// }
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
