package entity

import (
	"context"

	"github.com/bruno303/go-toolkit/pkg/log"
)

type (
	Client struct {
		ID   string
		Name string

		room *Room

		vote        *string
		HasVoted    bool
		IsSpectator bool
		IsOwner     bool

		logger log.Logger
	}
)

func newClient(id string) *Client {
	return &Client{
		ID:     id,
		logger: log.NewLogger("planningpoker.client"),
	}
}

func (c *Client) Vote(ctx context.Context, vote *string) {
	if c.room.Reveal {
		c.logger.Debug(ctx, "Vote ignored for client %s because votes are already revealed", c.ID)
		return
	}

	c.vote = vote
	if vote != nil && *vote != "" {
		c.HasVoted = true
	} else {
		c.HasVoted = false
	}

	c.room.checkReveal()
}

func (c *Client) Room() *Room {
	return c.room
}

func (c *Client) GetVote() *string {
	return c.vote
}

func (c *Client) UpdateName(ctx context.Context, name string) {
	c.Name = name
}
