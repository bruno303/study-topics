package repository

import (
	"context"

	"github.com/bruno303/study-topics/go-study/internal/application/hello/models"
	applicationRepository "github.com/bruno303/study-topics/go-study/internal/application/repository"
	"github.com/bruno303/study-topics/go-study/internal/infra/database"
)

type HelloMemDbRepository struct {
	db *database.MemDbRepository[models.HelloData]
}

var _ applicationRepository.HelloRepository = (*HelloMemDbRepository)(nil)

func NewHelloMemDbRepository(db *database.MemDbRepository[models.HelloData]) HelloMemDbRepository {
	return HelloMemDbRepository{db: db}
}

func (r HelloMemDbRepository) Save(ctx context.Context, entity *models.HelloData) (*models.HelloData, error) {
	return r.db.Save(ctx, entity)
}

func (r HelloMemDbRepository) ListAll(ctx context.Context) ([]models.HelloData, error) {
	return r.db.ListAll(ctx), nil
}
