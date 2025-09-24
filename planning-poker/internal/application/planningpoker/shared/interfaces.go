package shared

import (
	"context"
	"planning-poker/internal/application/planningpoker/entity"
)

type (
	Hub interface {
		FindClientByID(clientID string) (*entity.Client, bool)
		AddClient(c *entity.Client)
		RemoveClient(ctx context.Context, clientID string, roomID string)

		NewRoom(owner string) *entity.Room
		GetRoom(roomID string) (*entity.Room, bool)
		RemoveRoom(roomID string)
		BroadcastToRoom(ctx context.Context, roomID string, message any) error

		GetBus(clientID string) (Bus, bool)
		AddBus(clientID string, bus Bus)
		RemoveBus(clientID string)
	}

	Bus interface {
		Close() error
		Send(ctx context.Context, message any) error
		Listen(ctx context.Context)
	}
)
