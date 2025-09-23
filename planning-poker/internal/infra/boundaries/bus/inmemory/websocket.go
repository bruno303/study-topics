package inmemory

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"planning-poker/internal/application/planningpoker"

	"github.com/bruno303/go-toolkit/pkg/log"
	"github.com/gorilla/websocket"
)

type WebsocketBus struct {
	ID     string
	conn   *websocket.Conn
	logger log.Logger
}

func NewWebsocketBus(id string, socket *websocket.Conn) *WebsocketBus {
	return &WebsocketBus{
		ID:     id,
		conn:   socket,
		logger: log.NewLogger("websocket.client"),
	}
}

func (c *WebsocketBus) Close() error {
	return c.conn.Close()
}

func (c *WebsocketBus) Send(ctx context.Context, message any) error {
	c.logger.Debug(ctx, "Sending message to client: %v", message)
	return c.conn.WriteJSON(message)
}

func (c *WebsocketBus) receive() (map[string]any, error) {
	var msg map[string]any
	err := c.conn.ReadJSON(&msg)
	return msg, err
}

func (c *WebsocketBus) Listen(ctx context.Context, handleMessage func(msg planningpoker.Event)) {
	defer c.Close()

	for {
		msg, err := c.receive()
		if err != nil {
			if _, ok := err.(*websocket.CloseError); ok {
				c.logger.Info(ctx, "Client %v disconnected", c)

			} else {
				c.logger.Error(ctx, fmt.Sprintf("Error receiving message from client %v", c.ID), err)
			}

			return
		}
		c.logger.Info(ctx, "Message received from client %v: %v", c, msg)

		eventType, ok := msg["type"].(string)
		if !ok {
			c.logger.Error(ctx, fmt.Sprintf("Error casting message type to string for client %v", c.ID), err)
			continue
		}

		jsonData, err := json.Marshal(msg)
		if err != nil {
			c.logger.Error(ctx, fmt.Sprintf("Error marshaling message to JSON for client %v", c.ID), err)
			continue
		}

		var e planningpoker.Event
		var uerr error
		switch eventType {
		case "init":
			e, uerr = unmarshalEvent[planningpoker.InitEvent](jsonData)
		case "vote":
			e, uerr = unmarshalEvent[planningpoker.VoteEvent](jsonData)
		case "reset":
			e, uerr = unmarshalEvent[planningpoker.ResetEvent](jsonData)
		case "reveal-votes":
			e, uerr = unmarshalEvent[planningpoker.RevealEvent](jsonData)
		case "toggle-spectator":
			e, uerr = unmarshalEvent[planningpoker.SpectatorEvent](jsonData)
		case "toggle-owner":
			e, uerr = unmarshalEvent[planningpoker.OwnerEvent](jsonData)
		case "update-story":
			e, uerr = unmarshalEvent[planningpoker.StoryEvent](jsonData)
		case "new-voting":
			e, uerr = unmarshalEvent[planningpoker.NewVotingEvent](jsonData)
		case "vote-again":
			e, uerr = unmarshalEvent[planningpoker.VoteAgainEvent](jsonData)
		default:
			c.logger.Error(ctx, fmt.Sprintf("Unknown event type '%v' for client %v", eventType, c.ID), errors.New("unknown event type"))
			continue
		}

		if uerr != nil {
			c.logger.Error(ctx, fmt.Sprintf("Error unmarshaling event for client %v", c.ID), uerr)
			continue
		}

		handleMessage(e)
	}
}

func unmarshalEvent[T planningpoker.Event](data []byte) (T, error) {
	var event T
	err := json.Unmarshal(data, &event)
	return event, err
}
