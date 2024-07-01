package handlers

import "context"

type MessageHandler interface {
	Process(ctx context.Context, msg string) error
}
