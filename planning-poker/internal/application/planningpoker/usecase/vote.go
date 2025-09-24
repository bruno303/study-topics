package usecase

import (
	"context"
	"errors"
	"planning-poker/internal/application/planningpoker/shared"
	"planning-poker/internal/application/planningpoker/usecase/dto"
)

type (
	VoteCommand struct {
		ClientID string
		Vote     *string
	}
	VoteUseCase struct {
		hub shared.Hub
	}
)

func NewVoteUseCase(hub shared.Hub) VoteUseCase {
	return VoteUseCase{
		hub: hub,
	}
}

func (uc VoteUseCase) Execute(ctx context.Context, cmd VoteCommand) error {
	client, ok := uc.hub.FindClientByID(cmd.ClientID)
	if !ok {
		return errors.New("client not found")
	}
	client.Vote(ctx, cmd.Vote)
	uc.hub.BroadcastToRoom(ctx, client.Room().ID, dto.NewRoomStateCommand(client.Room()))
	return nil
}
