package main

import (
	"planning-poker/internal/application/planningpoker/interfaces"
	"planning-poker/internal/infra/boundaries/bus/inmemory"
)

type Container struct {
	Hub        *inmemory.InMemoryHub
	BusFactory interfaces.BusFactory
}

func NewContainer() *Container {
	return &Container{
		Hub:        inmemory.NewHub(),
		BusFactory: inmemory.NewWebSocketBusFactory(),
	}
}
