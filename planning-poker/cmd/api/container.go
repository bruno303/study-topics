package main

import (
	"planning-poker/internal/application/planningpoker"
	"planning-poker/internal/infra/boundaries/bus/inmemory"
)

type Container struct {
	Hub                  *inmemory.InMemoryHub
	BusFactory           inmemory.WebSocketBusFactory
	EventHandlerStrategy planningpoker.EventHandlerStrategy
}

func NewContainer() *Container {
	eventHandlerStrategy := planningpoker.NewEventhandlerStrategyImpl(
		planningpoker.NewInitEventHandler(),
		planningpoker.NewVoteEventHandler(),
		planningpoker.NewRevealEventHandler(),
		planningpoker.NewResetEventHandler(),
		planningpoker.NewSpectatorEventHandler(),
		planningpoker.NewStoryEventHandler(),
		planningpoker.NewOwnerEventHandler(),
		planningpoker.NewNewVotingEventHandler(),
		planningpoker.NewVoteAgainEventHandler(),
	)

	return &Container{
		Hub:                  inmemory.NewHub(eventHandlerStrategy),
		BusFactory:           inmemory.NewWebSocketBusFactory(),
		EventHandlerStrategy: eventHandlerStrategy,
	}
}
