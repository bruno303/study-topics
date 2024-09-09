package hello

import (
	"context"
	"fmt"

	"github.com/bruno303/study-topics/go-study/internal/crosscutting/observability/trace"
	"github.com/bruno303/study-topics/go-study/internal/hello/hellomodel"
	"github.com/bruno303/study-topics/go-study/internal/unitofwork"
)

const traceName = "HelloService"

type HelloService struct {
	uow unitofwork.UnitOfWork
}

func NewService(uow unitofwork.UnitOfWork) HelloService {
	return HelloService{uow}
}

func (s HelloService) Hello(ctx context.Context, id string, age int) string {
	ctx, end := trace.Trace(ctx, trace.NameConfig(traceName, "Hello2"))
	defer end()

	tx, err := s.uow.BeginTransaction(ctx)
	if err != nil {
		trace.InjectError(ctx, err)
		return err.Error()
	}

	defer tx.Rollback(ctx)
	var repository hellomodel.HelloRepository = tx.HelloRepository()

	newHello := hellomodel.HelloData{Id: id, Name: fmt.Sprintf("Bruno %v", id), Age: age}
	helloAdded, err := repository.Save(ctx, &newHello)
	if err != nil {
		trace.InjectError(ctx, err)
		return err.Error()
	}
	helloFound, err := repository.FindById(ctx, helloAdded.Id)
	if err != nil {
		trace.InjectError(ctx, err)
		return err.Error()
	}
	tx.HelloFileService().WriteFile(ctx, helloFound.Name)
	if err = tx.Commit(ctx); err != nil {
		return err.Error()
	}
	return helloFound.ToString()
}

func (s HelloService) ListAll(ctx context.Context) ([]hellomodel.HelloData, error) {
	ctx, end := trace.Trace(ctx, trace.NameConfig(traceName, "ListAll"))
	defer end()

	uow, err := s.uow.BeginTransaction(ctx)
	if err != nil {
		return nil, err
	}
	return uow.HelloRepository().ListAll(ctx), nil
}

// func (s HelloService) Hello(ctx context.Context, id string, age int) string {
// 	ctx, end := trace.Trace(ctx, trace.NameConfig(traceName, "Hello"))
// 	defer end()

// 	data, err := s.repository.RunWithTransaction(ctx, func(ctxTx context.Context) (*HelloData, error) {

// 		ctxTx, end := trace.Trace(ctxTx, trace.NameConfig(traceName, "HelloTransactional"))
// 		defer end()

// 		newHello := HelloData{Id: id, Name: fmt.Sprintf("Bruno %v", id), Age: age}
// 		helloAdded, err := s.repository.Save(ctxTx, &newHello)
// 		if err != nil {
// 			trace.InjectError(ctxTx, err)
// 			return nil, err
// 		}
// 		helloFound, err := s.repository.FindById(ctxTx, helloAdded.Id)
// 		if err != nil {
// 			trace.InjectError(ctxTx, err)
// 			return nil, err
// 		}
// 		return helloFound, nil
// 	})
// 	if err != nil {
// 		trace.InjectError(ctx, err)
// 		return err.Error()
// 	}
// 	return data.ToString()
// }
