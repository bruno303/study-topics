package usecase

import (
	"context"
	"errors"
	"planning-poker/internal/application/planningpoker/shared"
	"planning-poker/internal/application/planningpoker/usecase/dto"
)

type (
	UpdateStoryCommand struct {
		ClientID string
		Story    string
	}
	UpdateStoryUseCase struct {
		hub shared.Hub
	}
)

func NewUpdateStoryUseCase(hub shared.Hub) UpdateStoryUseCase {
	return UpdateStoryUseCase{
		hub: hub,
	}
}

func (uc UpdateStoryUseCase) Execute(ctx context.Context, cmd UpdateStoryCommand) error {
	client, ok := uc.hub.FindClientByID(cmd.ClientID)
	if !ok {
		return errors.New("client not found")
	}
	if !client.IsOwner {
		return errors.New("only the room owner can update the story")
	}

	client.Room().SetCurrentStory(cmd.Story)
	uc.hub.BroadcastToRoom(ctx, client.Room().ID, dto.NewRoomStateCommand(client.Room()))

	return nil
}
