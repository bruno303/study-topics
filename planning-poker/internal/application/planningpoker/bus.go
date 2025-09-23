package planningpoker

import (
	"context"
)

type (
	Bus interface {
		Close() error
		Listen(ctx context.Context, handleMessage func(msg Event))
		Send(ctx context.Context, message any) error
	}

	BusFactory interface {
		Create(ctx context.Context, id string, connection any) (Bus, error)
	}
)
