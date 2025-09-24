package usecase

import (
	"context"
	"errors"
	"planning-poker/internal/application/planningpoker/shared"
	"planning-poker/internal/application/planningpoker/usecase/dto"
)

type (
	ToggleSpectatorCommand struct {
		ClientID       string
		TargetClientID string
	}
	ToggleSpectatorUseCase struct {
		hub shared.Hub
	}
)

func NewToggleSpectatorUseCase(hub shared.Hub) ToggleSpectatorUseCase {
	return ToggleSpectatorUseCase{
		hub: hub,
	}
}

func (uc ToggleSpectatorUseCase) Execute(ctx context.Context, cmd ToggleSpectatorCommand) error {
	client, ok := uc.hub.FindClientByID(cmd.ClientID)
	if !ok {
		return errors.New("client not found")
	}

	if !client.IsOwner {
		return errors.New("only the room owner can toggle spectator mode")
	}

	client.Room().ToggleSpectator(ctx, cmd.TargetClientID)
	uc.hub.BroadcastToRoom(ctx, client.Room().ID, dto.NewRoomStateCommand(client.Room()))

	return nil
}
