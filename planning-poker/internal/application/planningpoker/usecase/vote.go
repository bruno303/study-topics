package usecase

import (
	"context"
	"fmt"
	"planning-poker/internal/application/planningpoker/usecase/dto"
	"planning-poker/internal/domain"
)

type (
	VoteCommand struct {
		RoomID   string
		SenderID string
		Vote     *string
	}
	VoteUseCase struct {
		hub domain.Hub
	}
)

func NewVoteUseCase(hub domain.Hub) VoteUseCase {
	return VoteUseCase{
		hub: hub,
	}
}

func (uc VoteUseCase) Execute(ctx context.Context, cmd VoteCommand) error {
	room, ok := uc.hub.GetRoom(cmd.RoomID)
	if !ok {
		return fmt.Errorf("room %s not found", cmd.RoomID)
	}

	if err := room.Vote(ctx, cmd.SenderID, cmd.Vote); err != nil {
		return err
	}

	if err := uc.hub.BroadcastToRoom(ctx, room.ID, dto.NewRoomStateCommand(room)); err != nil {
		return err
	}

	return nil
}
