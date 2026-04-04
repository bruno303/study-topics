package hello

import (
	"context"
	"fmt"

	"github.com/bruno303/study-topics/go-study/internal/application/hello/models"
	"github.com/bruno303/study-topics/go-study/internal/application/transaction"
	"github.com/google/uuid"
)

//go:generate go tool mockgen -source=service.go -destination=mocks.go -package hello

type (
	HelloService interface {
		ListAll(ctx context.Context) ([]models.HelloData, error)
		Hello(ctx context.Context, input HelloInput) (models.HelloData, error)
	}

	HelloInput struct {
		Id  string
		Age int
	}
)

var _ HelloService = (*helloService)(nil)

type helloService struct {
	transactionManager transaction.TransactionManager
}

func NewService(transactionManager transaction.TransactionManager) helloService {
	return helloService{transactionManager: transactionManager}
}

func (s helloService) ListAll(ctx context.Context) ([]models.HelloData, error) {
	var result []models.HelloData

	err := s.transactionManager.WithinTx(ctx, transaction.EmptyOpts, func(txCtx context.Context, uow transaction.UnitOfWork) error {
		list, err := uow.HelloRepository().ListAll(txCtx)
		if err != nil {
			return err
		}

		result = list
		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s helloService) Hello(ctx context.Context, input HelloInput) (models.HelloData, error) {
	var hello1 models.HelloData

	err := s.transactionManager.WithinTx(ctx, transaction.EmptyOpts, func(txCtx context.Context, uow transaction.UnitOfWork) error {
		hello1 = models.HelloData{Id: input.Id, Name: fmt.Sprintf("Bruno %v", input.Id), Age: input.Age}
		_, err := uow.HelloRepository().Save(txCtx, &hello1)
		if err != nil {
			return err
		}

		id2 := uuid.NewString()
		hello2 := models.HelloData{Id: id2, Name: fmt.Sprintf("Bruno %v", id2), Age: input.Age}
		_, err = uow.HelloRepository().Save(txCtx, &hello2)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return models.HelloData{}, err
	}

	return hello1, nil
}
