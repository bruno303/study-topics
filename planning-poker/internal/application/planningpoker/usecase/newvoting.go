package usecase

import (
	"context"
	"fmt"
	"planning-poker/internal/application/planningpoker/usecase/dto"
	"planning-poker/internal/domain"
)

type (
	NewVotingCommand struct {
		RoomID   string
		SenderID string
	}
	NewVotingUseCase struct {
		hub domain.Hub
	}
)

func NewNewVotingUseCase(hub domain.Hub) NewVotingUseCase {
	return NewVotingUseCase{
		hub: hub,
	}
}

func (uc NewVotingUseCase) Execute(ctx context.Context, cmd NewVotingCommand) error {
	room, ok := uc.hub.GetRoom(cmd.RoomID)
	if !ok {
		return fmt.Errorf("room %s not found", cmd.RoomID)
	}

	if err := room.NewVoting(ctx, cmd.SenderID); err != nil {
		return err
	}

	if err := uc.hub.BroadcastToRoom(ctx, room.ID, dto.NewRoomStateCommand(room)); err != nil {
		return err
	}

	return nil
}
