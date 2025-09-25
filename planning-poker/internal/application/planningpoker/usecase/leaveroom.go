package usecase

import (
	"context"
	"planning-poker/internal/application/planningpoker/usecase/dto"
	"planning-poker/internal/domain"
)

type (
	LeaveRoomCommand struct {
		RoomID   string
		SenderID string
	}
	LeaveRoomUseCase struct {
		hub domain.Hub
	}
)

func NewLeaveRoomUseCase(hub domain.Hub) LeaveRoomUseCase {
	return LeaveRoomUseCase{
		hub: hub,
	}
}

func (uc LeaveRoomUseCase) Execute(ctx context.Context, cmd LeaveRoomCommand) error {
	if err := uc.hub.RemoveClient(ctx, cmd.SenderID, cmd.RoomID); err != nil {
		return err
	}

	// if room still exists, broadcast the updated state
	if room, ok := uc.hub.GetRoom(cmd.RoomID); ok {
		if err := uc.hub.BroadcastToRoom(ctx, room.ID, dto.NewRoomStateCommand(room)); err != nil {
			return err
		}
	}

	return nil
}
