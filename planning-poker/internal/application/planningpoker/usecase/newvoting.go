package usecase

import (
	"context"
	"errors"
	"planning-poker/internal/application/planningpoker/shared"
	"planning-poker/internal/application/planningpoker/usecase/dto"
)

type (
	NewVotingCommand struct {
		ClientID string
	}
	NewVotingUseCase struct {
		hub shared.Hub
	}
)

func NewNewVotingUseCase(hub shared.Hub) NewVotingUseCase {
	return NewVotingUseCase{
		hub: hub,
	}
}

func (uc NewVotingUseCase) Execute(ctx context.Context, cmd NewVotingCommand) error {
	client, ok := uc.hub.FindClientByID(cmd.ClientID)
	if !ok {
		return errors.New("client not found")
	}
	if !client.IsOwner {
		return errors.New("only the room owner can update start a new voting")
	}

	client.Room().NewVoting()

	uc.hub.BroadcastToRoom(ctx, client.Room().ID, dto.NewRoomStateCommand(client.Room()))

	return nil
}
