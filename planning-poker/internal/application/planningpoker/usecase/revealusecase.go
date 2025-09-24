package usecase

import (
	"context"
	"errors"
	"planning-poker/internal/application/planningpoker/shared"
	"planning-poker/internal/application/planningpoker/usecase/dto"
)

type (
	RevealCommand struct {
		ClientID string
	}
	RevealUseCase struct {
		hub shared.Hub
	}
)

func NewRevealUseCase(hub shared.Hub) RevealUseCase {
	return RevealUseCase{
		hub: hub,
	}
}

func (uc RevealUseCase) Execute(ctx context.Context, cmd RevealCommand) error {
	client, ok := uc.hub.FindClientByID(cmd.ClientID)
	if !ok {
		return errors.New("client not found")
	}

	if !client.IsOwner {
		return errors.New("only the room owner can toggle reveal")
	}

	room, ok := uc.hub.GetRoom(client.Room().ID)
	if !ok {
		return errors.New("room not found")
	}

	room.ToggleReveal()

	uc.hub.BroadcastToRoom(ctx, client.Room().ID, dto.NewRoomStateCommand(client.Room()))

	return nil
}
