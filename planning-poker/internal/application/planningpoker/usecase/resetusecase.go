package usecase

import (
	"context"
	"fmt"
	"planning-poker/internal/application/planningpoker/usecase/dto"
	"planning-poker/internal/domain"
)

type (
	ResetCommand struct {
		RoomID   string
		SenderID string
	}
	ResetUseCase struct {
		hub domain.Hub
	}
)

func NewResetUseCase(hub domain.Hub) ResetUseCase {
	return ResetUseCase{
		hub: hub,
	}
}

func (uc ResetUseCase) Execute(ctx context.Context, cmd ResetCommand) error {
	room, ok := uc.hub.GetRoom(cmd.RoomID)
	if !ok {
		return fmt.Errorf("room %s not found", cmd.RoomID)
	}

	if err := room.ResetVoting(ctx, cmd.SenderID); err != nil {
		return err
	}

	if err := uc.hub.BroadcastToRoom(ctx, room.ID, dto.NewRoomStateCommand(room)); err != nil {
		return err
	}

	return nil
}
