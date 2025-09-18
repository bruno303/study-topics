package bus

import (
	"context"

	"github.com/bruno303/go-toolkit/pkg/log"
	"github.com/google/uuid"
)

type InMemoryRoom struct {
	ID      string
	clients []*WebsocketBus
	logger  log.Logger
	Hub     *InMemoryHub
}

func NewRoom() *InMemoryRoom {
	return &InMemoryRoom{
		ID:      uuid.New().String(),
		clients: make([]*WebsocketBus, 0),
		logger:  log.NewLogger("websocket.room"),
	}
}

func (r *InMemoryRoom) AddClient(client *WebsocketBus) {
	client.room = r
	r.clients = append(r.clients, client)
}

func (r *InMemoryRoom) RemoveClient(client *WebsocketBus) {
	for i, c := range r.clients {
		if c == client {
			r.clients = append(r.clients[:i], r.clients[i+1:]...)
			return
		}
	}
	if len(r.clients) == 0 {
		r.Hub.RemoveRoom(r.ID)
		r.logger.Info(context.Background(), "Room '%s' removed from hub", r.ID)
	}
}

func (r *InMemoryRoom) Broadcast(ctx context.Context, message any) {
	for _, client := range r.clients {
		err := client.conn.WriteJSON(message)
		if err != nil {
			r.logger.Error(ctx, "Error broadcasting message to client: %v", err)
		}
	}
}
