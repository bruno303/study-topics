package repository

import (
	"context"

	"github.com/bruno303/study-topics/go-study/internal/application/hello/models"
)

type HelloRepository interface {
	Save(ctx context.Context, entity *models.HelloData) (*models.HelloData, error)
	ListAll(ctx context.Context) ([]models.HelloData, error)
}
