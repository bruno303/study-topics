package main

import "planning-poker/internal/infra/boundaries/bus"

type Container struct {
	Hub *bus.InMemoryHub
}

func NewContainer() *Container {
	return &Container{
		Hub: bus.NewHub(),
	}
}
