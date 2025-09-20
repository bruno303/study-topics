package inmemory

import (
	"context"
	"planning-poker/internal/application/planningpoker/interfaces"

	"github.com/gorilla/websocket"
)

type WebSocketBusFactory struct{}

func NewWebSocketBusFactory() WebSocketBusFactory {
	return WebSocketBusFactory{}
}

func (f WebSocketBusFactory) Create(ctx context.Context, id string, connection any) (interfaces.Bus, error) {
	return NewWebsocketBus(id, connection.(*websocket.Conn)), nil
}
