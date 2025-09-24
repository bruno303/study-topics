package usecase

import (
	"context"
	"errors"
	"planning-poker/internal/application/planningpoker/shared"
	"planning-poker/internal/application/planningpoker/usecase/dto"
)

type (
	ResetCommand struct {
		ClientID string
	}
	ResetUseCase struct {
		hub shared.Hub
	}
)

func NewResetUseCase(hub shared.Hub) ResetUseCase {
	return ResetUseCase{
		hub: hub,
	}
}

func (uc ResetUseCase) Execute(ctx context.Context, cmd ResetCommand) error {
	client, ok := uc.hub.FindClientByID(cmd.ClientID)
	if !ok {
		return errors.New("client not found")
	}

	if !client.IsOwner {
		return errors.New("only the room owner can reset voting")
	}

	room, ok := uc.hub.GetRoom(client.Room().ID)
	if !ok {
		return errors.New("room not found")
	}

	room.ResetVoting()

	uc.hub.BroadcastToRoom(ctx, client.Room().ID, dto.NewRoomStateCommand(client.Room()))

	return nil
}
