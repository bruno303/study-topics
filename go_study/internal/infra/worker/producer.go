package worker

import "context"

type Producer interface {
	Produce(ctx context.Context, msg string, topic string) error
	Close()
}
