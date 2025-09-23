package planningpoker

import (
	"context"

	"github.com/bruno303/go-toolkit/pkg/log"
)

type Client struct {
	ID   string
	Name string

	room *Room

	Vote        *string
	HasVoted    bool
	IsSpectator bool
	IsOwner     bool

	bus             Bus
	handlerStrategy EventHandlerStrategy

	logger log.Logger
}

func newClient(id string, bus Bus, handlerStrategy EventHandlerStrategy) *Client {
	return &Client{
		ID:              id,
		bus:             bus,
		handlerStrategy: handlerStrategy,
		logger:          log.NewLogger("planningpoker.client"),
	}
}

func (c *Client) Close() error {
	return c.bus.Close()
}

func (c *Client) Send(ctx context.Context, message any) error {
	log.Log().Debug(ctx, "Sending message to client %v: %v", c.ID, message)
	return c.bus.Send(ctx, message)
}

func (c *Client) vote(ctx context.Context, vote *string) {
	if c.room.Reveal {
		c.logger.Debug(ctx, "Vote ignored for client %s because votes are already revealed", c.ID)
		return
	}

	c.Vote = vote
	if vote != nil && *vote != "" {
		c.HasVoted = true
	} else {
		c.HasVoted = false
	}

	c.room.checkReveal()
}

func (c *Client) updateName(ctx context.Context, name string) {
	c.Name = name
	c.Send(ctx, NewUpdateClientIDCommand(c.ID))
}

func (c *Client) Listen(ctx context.Context) {
	c.Send(ctx, NewUpdateClientIDCommand(c.ID))

	go c.bus.Listen(ctx, func(msg Event) {
		c.handlerStrategy.HandleEvent(ctx, msg, c)
		c.room.Broadcast(ctx, NewRoomStateCommand(c.room))
	})
}
