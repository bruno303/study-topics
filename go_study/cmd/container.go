package main

import (
	"context"
	"main/internal/config"
	"main/internal/hello"
	"main/internal/infra"
)

type Container struct {
	Config *config.Config
	Hello  hello.Container
	Infra  infra.Container
}

func NewContainer(ctx context.Context) *Container {
	config := config.LoadConfig()
	infraContainer := infra.NewContainer(config)
	helloContainer := hello.NewContainer(ctx, infraContainer)

	return &Container{
		Config: config,
		Hello:  helloContainer,
		Infra:  infraContainer,
	}
}
