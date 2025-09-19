package interfaces

import (
	"context"
	"planning-poker/internal/infra/boundaries/bus/events"
)

type (
	Bus interface {
		Close() error
		Listen(ctx context.Context, handleMessage func(msg events.Event))
		Send(ctx context.Context, message any) error
	}

	BusFactory interface {
		Create(ctx context.Context, id string, connection any) (Bus, error)
	}
)
