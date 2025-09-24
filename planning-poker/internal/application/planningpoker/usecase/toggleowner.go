package usecase

import (
	"context"
	"errors"
	"planning-poker/internal/application/planningpoker/shared"
	"planning-poker/internal/application/planningpoker/usecase/dto"
)

type (
	ToggleOwnerCommand struct {
		ClientID       string
		TargetClientID string
	}
	ToggleOwnerUseCase struct {
		hub shared.Hub
	}
)

func NewToggleOwnerUseCase(hub shared.Hub) ToggleOwnerUseCase {
	return ToggleOwnerUseCase{
		hub: hub,
	}
}

func (uc ToggleOwnerUseCase) Execute(ctx context.Context, cmd ToggleOwnerCommand) error {
	client, ok := uc.hub.FindClientByID(cmd.ClientID)
	if !ok {
		return errors.New("client not found")
	}

	if !client.IsOwner {
		return errors.New("only the room owner can toggle ownership")
	}

	client.Room().ToggleOwner(cmd.TargetClientID)
	uc.hub.BroadcastToRoom(ctx, client.Room().ID, dto.NewRoomStateCommand(client.Room()))

	return nil
}
