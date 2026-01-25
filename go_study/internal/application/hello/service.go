package hello

import (
	"context"
	"fmt"

	"github.com/bruno303/study-topics/go-study/internal/application/hello/models"
	"github.com/bruno303/study-topics/go-study/internal/application/transaction"
	"github.com/bruno303/study-topics/go-study/pkg/utils/generic"
	"github.com/google/uuid"
)

//go:generate go tool mockgen -source=service.go -destination=mocks.go -package hello

type HelloRepository interface {
	Save(ctx context.Context, entity *models.HelloData, tx transaction.Transaction) (*models.HelloData, error)
	ListAll(ctx context.Context, tx transaction.Transaction) []models.HelloData
}

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
	helloRepository HelloRepository
	txManager       transaction.TransactionManager
}

func NewService(
	helloRepository HelloRepository,
	txManager transaction.TransactionManager,
) helloService {
	return helloService{helloRepository, txManager}
}

func (s helloService) ListAll(ctx context.Context) ([]models.HelloData, error) {
	result, _ := s.txManager.Execute(ctx, transaction.EmptyOpts, func(ctx context.Context, tx transaction.Transaction) (any, error) {
		return s.helloRepository.ListAll(ctx, tx), nil
	})

	return result.([]models.HelloData), nil
}

func (s helloService) Hello(ctx context.Context, input HelloInput) (models.HelloData, error) {
	result, err := s.txManager.Execute(ctx, transaction.EmptyOpts, func(ctx context.Context, tx transaction.Transaction) (any, error) {
		var result models.HelloData

		hello1 := models.HelloData{Id: input.Id, Name: fmt.Sprintf("Bruno %v", input.Id), Age: input.Age}
		_, err := s.helloRepository.Save(ctx, &hello1, tx)
		if err != nil {
			return result, err
		}

		id2 := uuid.NewString()
		hello2 := models.HelloData{Id: id2, Name: fmt.Sprintf("Bruno %v", id2), Age: input.Age}
		_, err = s.helloRepository.Save(ctx, &hello2, tx)
		if err != nil {
			return result, err
		}

		return hello1, nil
	})

	return generic.CastValueIfNoError[models.HelloData](result, err)
}
