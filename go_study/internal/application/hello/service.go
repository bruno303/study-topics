package hello

import (
	"context"
	"fmt"

	"github.com/bruno303/study-topics/go-study/internal/application/hello/models"
	"github.com/google/uuid"
)

//go:generate mockgen -source=service.go -destination=mocks.go -package hello

type HelloRepository interface {
	Save(ctx context.Context, entity *models.HelloData) (*models.HelloData, error)
	FindById(ctx context.Context, id any) (*models.HelloData, error)
	ListAll(ctx context.Context) []models.HelloData
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
}

func NewService(
	helloRepository HelloRepository,
) helloService {
	return helloService{helloRepository}
}

func (s helloService) ListAll(ctx context.Context) ([]models.HelloData, error) {
	return s.helloRepository.ListAll(ctx), nil
}

func (s helloService) Hello(ctx context.Context, input HelloInput) (models.HelloData, error) {
	var result models.HelloData

	hello1 := models.HelloData{Id: input.Id, Name: fmt.Sprintf("Bruno %v", input.Id), Age: input.Age}
	_, err := s.helloRepository.Save(ctx, &hello1)
	if err != nil {
		return result, err
	}

	id2 := uuid.NewString()
	hello2 := models.HelloData{Id: id2, Name: fmt.Sprintf("Bruno %v", id2), Age: input.Age}
	_, err = s.helloRepository.Save(ctx, &hello2)
	if err != nil {
		return result, err
	}

	return hello1, nil
}
