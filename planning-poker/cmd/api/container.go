package main

import (
	"planning-poker/internal/application/planningpoker/usecase"
	"planning-poker/internal/infra/boundaries/bus/inmemory"
)

type Container struct {
	Hub      *inmemory.InMemoryHub
	Usecases usecase.UseCases
}

func NewContainer() *Container {
	hub := inmemory.NewHub()

	return &Container{
		Hub: hub,
		Usecases: usecase.UseCases{
			UpdateName:      usecase.NewUpdateNameUseCase(hub),
			Vote:            usecase.NewVoteUseCase(hub),
			Reveal:          usecase.NewRevealUseCase(hub),
			Reset:           usecase.NewResetUseCase(hub),
			ToggleSpectator: usecase.NewToggleSpectatorUseCase(hub),
			ToggleOwner:     usecase.NewToggleOwnerUseCase(hub),
			UpdateStory:     usecase.NewUpdateStoryUseCase(hub),
			NewVoting:       usecase.NewNewVotingUseCase(hub),
			VoteAgain:       usecase.NewVoteAgainUseCase(hub),
			LeaveRoom:       usecase.NewLeaveRoomUseCase(hub),
		},
	}
}
