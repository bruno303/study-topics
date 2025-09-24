package usecase

import (
	"context"
	"errors"
	"planning-poker/internal/application/planningpoker/shared"
	"planning-poker/internal/application/planningpoker/usecase/dto"
)

type (
	VoteAgainCommand struct {
		ClientID string
	}
	VoteAgainUseCase struct {
		hub shared.Hub
	}
)

func NewVoteAgainUseCase(hub shared.Hub) VoteAgainUseCase {
	return VoteAgainUseCase{
		hub: hub,
	}
}

func (uc VoteAgainUseCase) Execute(ctx context.Context, cmd VoteAgainCommand) error {
	client, ok := uc.hub.FindClientByID(cmd.ClientID)
	if !ok {
		return errors.New("client not found")
	}
	if !client.IsOwner {
		return errors.New("only the room owner can update the story")
	}

	client.Room().ResetVoting()
	uc.hub.BroadcastToRoom(ctx, client.Room().ID, dto.NewRoomStateCommand(client.Room()))

	return nil
}
