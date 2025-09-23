package planningpoker

import (
	"context"

	"github.com/samber/lo"
)

type (
	EventHandlerTypeAware struct {
		Type string `json:"type"`
	}

	InitEventHandler      struct{ EventHandlerTypeAware }
	VoteEventHandler      struct{ EventHandlerTypeAware }
	RevealEventHandler    struct{ EventHandlerTypeAware }
	ResetEventHandler     struct{ EventHandlerTypeAware }
	SpectatorEventHandler struct{ EventHandlerTypeAware }
	StoryEventHandler     struct{ EventHandlerTypeAware }
	OwnerEventHandler     struct{ EventHandlerTypeAware }
	NewVotingEventHandler struct{ EventHandlerTypeAware }
	VoteAgainEventHandler struct{ EventHandlerTypeAware }
)

func (h EventHandlerTypeAware) GetType() string { return h.Type }

func NewInitEventHandler() InitEventHandler {
	return InitEventHandler{EventHandlerTypeAware: EventHandlerTypeAware{Type: "init"}}
}

func NewVoteEventHandler() VoteEventHandler {
	return VoteEventHandler{EventHandlerTypeAware: EventHandlerTypeAware{Type: "vote"}}
}

func NewRevealEventHandler() RevealEventHandler {
	return RevealEventHandler{EventHandlerTypeAware: EventHandlerTypeAware{Type: "reveal-votes"}}
}

func NewResetEventHandler() ResetEventHandler {
	return ResetEventHandler{EventHandlerTypeAware: EventHandlerTypeAware{Type: "reset"}}
}

func NewSpectatorEventHandler() SpectatorEventHandler {
	return SpectatorEventHandler{EventHandlerTypeAware: EventHandlerTypeAware{Type: "toggle-spectator"}}
}

func NewStoryEventHandler() StoryEventHandler {
	return StoryEventHandler{EventHandlerTypeAware: EventHandlerTypeAware{Type: "update-story"}}
}

func NewOwnerEventHandler() OwnerEventHandler {
	return OwnerEventHandler{EventHandlerTypeAware: EventHandlerTypeAware{Type: "toggle-owner"}}
}

func NewNewVotingEventHandler() NewVotingEventHandler {
	return NewVotingEventHandler{EventHandlerTypeAware: EventHandlerTypeAware{Type: "new-voting"}}
}

func NewVoteAgainEventHandler() VoteAgainEventHandler {
	return VoteAgainEventHandler{EventHandlerTypeAware: EventHandlerTypeAware{Type: "vote-again"}}
}

func (h InitEventHandler) Handle(ctx context.Context, event any, client *Client) error {
	eventTyped, ok := event.(InitEvent)
	if !ok {
		return nil
	}
	client.updateName(ctx, eventTyped.Username)
	return nil
}

func (h VoteEventHandler) Handle(ctx context.Context, event any, client *Client) error {
	eventTyped, ok := event.(VoteEvent)
	if !ok {
		return nil
	}
	client.vote(ctx, lo.ToPtr(eventTyped.Vote))
	return nil
}

func (h RevealEventHandler) Handle(ctx context.Context, event any, client *Client) error {
	_, ok := event.(RevealEvent)
	if !ok {
		return nil
	}
	return executeIfOwner(client, func() error {
		client.room.ToggleReveal()
		return nil
	})
}

func (h ResetEventHandler) Handle(ctx context.Context, event any, client *Client) error {
	_, ok := event.(ResetEvent)
	if !ok {
		return nil
	}
	return executeIfOwner(client, func() error {
		client.room.ResetVoting()
		return nil
	})
}

func (h SpectatorEventHandler) Handle(ctx context.Context, event any, client *Client) error {
	eventTyped, ok := event.(SpectatorEvent)
	if !ok {
		return nil
	}
	return executeIfOwner(client, func() error {
		client.room.ToggleSpectator(ctx, eventTyped.ClientID)
		return nil
	})
}

func (h StoryEventHandler) Handle(ctx context.Context, event any, client *Client) error {
	eventTyped, ok := event.(StoryEvent)
	if !ok {
		return nil
	}
	return executeIfOwner(client, func() error {
		client.room.SetCurrentStory(eventTyped.Story)
		return nil
	})
}

func (h OwnerEventHandler) Handle(ctx context.Context, event any, client *Client) error {
	eventTyped, ok := event.(OwnerEvent)
	if !ok {
		return nil
	}
	return executeIfOwner(client, func() error {
		client.room.ToggleOwner(eventTyped.ClientID)
		return nil
	})
}

func (h NewVotingEventHandler) Handle(ctx context.Context, event any, client *Client) error {
	_, ok := event.(NewVotingEvent)
	if !ok {
		return nil
	}
	return executeIfOwner(client, func() error {
		client.room.NewVoting()
		return nil
	})
}

func (h VoteAgainEventHandler) Handle(ctx context.Context, event any, client *Client) error {
	_, ok := event.(VoteAgainEvent)
	if !ok {
		return nil
	}
	return executeIfOwner(client, func() error {
		client.room.ResetVoting()
		return nil
	})
}

var (
	_ EventHandler = (*InitEventHandler)(nil)
	_ EventHandler = (*VoteEventHandler)(nil)
	_ EventHandler = (*RevealEventHandler)(nil)
	_ EventHandler = (*ResetEventHandler)(nil)
	_ EventHandler = (*SpectatorEventHandler)(nil)
	_ EventHandler = (*StoryEventHandler)(nil)
	_ EventHandler = (*OwnerEventHandler)(nil)
	_ EventHandler = (*NewVotingEventHandler)(nil)
	_ EventHandler = (*VoteAgainEventHandler)(nil)
)

func executeIfOwner(client *Client, fn func() error) error {
	if client.IsOwner {
		return fn()
	}
	return nil
}
