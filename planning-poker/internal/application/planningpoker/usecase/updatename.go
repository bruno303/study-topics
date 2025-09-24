package usecase

import (
	"context"
	"errors"
	"planning-poker/internal/application/planningpoker/shared"
	"planning-poker/internal/application/planningpoker/usecase/dto"
)

type (
	UpdateNameCommand struct {
		ClientID string
		Username string
	}
	UpdateNameUseCase struct {
		hub shared.Hub
	}
)

func NewUpdateNameUseCase(hub shared.Hub) UpdateNameUseCase {
	return UpdateNameUseCase{
		hub: hub,
	}
}

func (uc UpdateNameUseCase) Execute(ctx context.Context, cmd UpdateNameCommand) error {
	client, ok := uc.hub.FindClientByID(cmd.ClientID)
	if !ok {
		return errors.New("client not found")
	}

	client.UpdateName(ctx, cmd.Username)
	uc.hub.BroadcastToRoom(ctx, client.Room().ID, dto.NewRoomStateCommand(client.Room()))

	return nil
}
