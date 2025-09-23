package inmemory

import (
	"context"
	"planning-poker/internal/application/planningpoker"

	"github.com/gorilla/websocket"
)

type WebSocketBusFactory struct{}

func NewWebSocketBusFactory() WebSocketBusFactory {
	return WebSocketBusFactory{}
}

func (f WebSocketBusFactory) Create(ctx context.Context, id string, connection any) (planningpoker.Bus, error) {
	return NewWebsocketBus(id, connection.(*websocket.Conn)), nil
}
