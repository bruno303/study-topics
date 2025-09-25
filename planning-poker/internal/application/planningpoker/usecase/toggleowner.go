package usecase

import (
	"context"
	"fmt"
	"planning-poker/internal/application/planningpoker/usecase/dto"
	"planning-poker/internal/domain"
)

type (
	ToggleOwnerCommand struct {
		RoomID         string
		SenderID       string
		TargetClientID string
	}
	ToggleOwnerUseCase struct {
		hub domain.Hub
	}
)

func NewToggleOwnerUseCase(hub domain.Hub) ToggleOwnerUseCase {
	return ToggleOwnerUseCase{
		hub: hub,
	}
}

func (uc ToggleOwnerUseCase) Execute(ctx context.Context, cmd ToggleOwnerCommand) error {
	room, ok := uc.hub.GetRoom(cmd.RoomID)
	if !ok {
		return fmt.Errorf("room %s not found", cmd.RoomID)
	}

	if err := room.ToggleOwner(ctx, cmd.SenderID, cmd.TargetClientID); err != nil {
		return err
	}

	if err := uc.hub.BroadcastToRoom(ctx, room.ID, dto.NewRoomStateCommand(room)); err != nil {
		return err
	}

	return nil
}
