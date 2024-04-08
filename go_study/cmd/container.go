package main

import (
	"context"
	"main/internal/config"
	"main/internal/hello"
	"main/internal/infra/database"
	"main/internal/infra/repository"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Container struct {
	Config       *config.Config
	Services     ServiceContainer
	Repositories RepositoryContainer
}

type ServiceContainer struct {
	HelloService hello.HelloService
}

type RepositoryContainer struct {
	HelloRepository hello.Repository
}

func newServiceContainer(repository hello.Repository) ServiceContainer {
	return ServiceContainer{
		HelloService: hello.NewService(repository),
	}
}

func newRepositoryContainer(ctx context.Context, pool *pgxpool.Pool) RepositoryContainer {
	return RepositoryContainer{
		HelloRepository: repository.NewHelloRepository(ctx, pool),
	}
}

func NewContainer(ctx context.Context) *Container {
	cfg := config.LoadConfig()
	pool := database.Connect(cfg)

	repositories := newRepositoryContainer(ctx, pool)
	services := newServiceContainer(repositories.HelloRepository)

	return &Container{
		Config:       cfg,
		Services:     services,
		Repositories: repositories,
	}
}
