package transactionaldecorator

import (
	"context"

	"github.com/bruno303/study-topics/go-study/internal/application/hello"
	"github.com/bruno303/study-topics/go-study/internal/application/hello/models"
	"github.com/bruno303/study-topics/go-study/internal/application/transaction"
)

type HelloServiceTransactionalDecorator struct {
	next    hello.HelloService
	manager transaction.TransactionManager
}

var _ hello.HelloService = (*HelloServiceTransactionalDecorator)(nil)

func NewTransactionalHelloService(next hello.HelloService, manager transaction.TransactionManager) HelloServiceTransactionalDecorator {
	return HelloServiceTransactionalDecorator{next, manager}
}

func (h HelloServiceTransactionalDecorator) Hello(ctx context.Context, input hello.HelloInput) (models.HelloData, error) {
	result, err := h.manager.Execute(ctx, func(txCtx context.Context) (any, error) {
		return h.next.Hello(txCtx, input)
	})
	if err != nil {
		return models.HelloData{}, err
	}
	return result.(models.HelloData), nil
}

func (h HelloServiceTransactionalDecorator) ListAll(ctx context.Context) ([]models.HelloData, error) {
	result, err := h.manager.Execute(ctx, func(txCtx context.Context) (any, error) {
		return h.next.ListAll(txCtx)
	})
	if err != nil {
		return nil, err
	}
	return result.([]models.HelloData), nil
}
