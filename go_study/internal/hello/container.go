package hello

import (
	"context"
	"main/internal/infra"
)

type Container struct {
	Repository Repository
	Service    HelloService
}

func NewContainer(ctx context.Context, container infra.Container) Container {
	helloRepository := NewRepository(ctx, container)
	helloService := NewService(helloRepository)

	return Container{
		Repository: helloRepository,
		Service:    helloService,
	}
}
