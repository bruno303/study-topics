package planningpoker

import (
	"context"
	"planning-poker/internal/application/planningpoker/interfaces"
	"planning-poker/internal/infra/boundaries/bus/events"

	"github.com/bruno303/go-toolkit/pkg/log"
	"github.com/samber/lo"
)

type Client struct {
	ID   string
	Name string

	room *Room

	Vote        *string
	HasVoted    bool
	IsSpectator bool
	IsOwner     bool

	bus interfaces.Bus

	logger log.Logger
}

func newClient(id string, bus interfaces.Bus) *Client {
	return &Client{
		ID:     id,
		bus:    bus,
		logger: log.NewLogger("planningpoker.client"),
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

	go c.bus.Listen(ctx, func(msg events.Event) {
		switch msg.Type() {
		case "init":
			c.updateName(ctx, msg.(events.InitEvent).Payload.Username)

		case "vote":
			c.vote(ctx, lo.ToPtr(msg.(events.VoteEvent).Payload.Vote))

		case "toggle-spectator":
			c.executeIfOwner(func() {
				c.room.ToggleSpectator(ctx, msg.(events.SpectatorEvent).Payload.ClientID)
			})

		case "toggle-owner":
			c.executeIfOwner(func() {
				c.room.ToggleOwner(msg.(events.OwnerEvent).Payload.ClientID)
			})

		case "new-voting":
			c.executeIfOwner(func() {
				c.room.NewVoting()
			})

		case "reveal-votes":
			c.executeIfOwner(func() {
				c.room.Reveal = !c.room.Reveal
			})

		case "update-story":
			c.executeIfOwner(func() {
				c.room.SetCurrentStory(msg.(events.StoryEvent).Payload.Story)
			})

		case "vote-again":
			c.executeIfOwner(func() {
				c.room.ResetVoting()
			})
		}

		// always broadcast the current state
		c.room.Broadcast(ctx, NewRoomStateCommand(c.room))
	})
}

func (c *Client) executeIfOwner(fn func()) {
	if c.IsOwner {
		fn()
	}
}
