package planningpoker

import (
	"context"

	"github.com/bruno303/go-toolkit/pkg/log"
)

type (
	Event interface {
		GetType() string
	}

	EventTypeAware struct {
		Type string `json:"type"`
	}

	EventHandler interface {
		Handle(ctx context.Context, event any, client *Client) error
		GetType() string
	}

	EventHandlerStrategy interface {
		HandleEvent(ctx context.Context, event any, client *Client) error
	}

	EventHandlerStrategyImpl struct {
		Handlers map[string]EventHandler
		logger   log.Logger
	}

	InitEvent struct {
		EventTypeAware
		Username string `json:"username"`
	}

	VoteEvent struct {
		EventTypeAware
		Vote string `json:"vote"`
	}

	RevealEvent struct {
		EventTypeAware
	}

	ResetEvent struct {
		EventTypeAware
	}

	SpectatorEvent struct {
		EventTypeAware
		ClientID string `json:"clientId"`
	}

	StoryEvent struct {
		EventTypeAware
		Story string `json:"story"`
	}

	OwnerEvent struct {
		EventTypeAware
		ClientID string `json:"clientId"`
	}

	NewVotingEvent struct {
		EventTypeAware
	}

	VoteAgainEvent struct {
		EventTypeAware
	}
)

func (e EventTypeAware) GetType() string { return e.Type }

func NewEventhandlerStrategyImpl(eventHandlers ...any) EventHandlerStrategyImpl {
	eventHandlerStrategy := EventHandlerStrategyImpl{
		Handlers: make(map[string]EventHandler),
		logger:   log.NewLogger("planningpoker.eventhandlerstrategy"),
	}
	for _, handler := range eventHandlers {
		switch h := handler.(type) {
		case EventHandler:
			eventHandlerStrategy.RegisterHandler(h.GetType(), h)
		default:
			eventHandlerStrategy.logger.Warn(context.Background(), "Invalid event handler type")
		}
	}
	return eventHandlerStrategy
}

func (s EventHandlerStrategyImpl) RegisterHandler(eventType string, handler EventHandler) {
	s.Handlers[eventType] = handler
}

func (s EventHandlerStrategyImpl) HandleEvent(ctx context.Context, event any, client *Client) error {
	eventTyped, ok := event.(Event)
	if !ok {
		s.logger.Warn(ctx, "Event does not implement Event interface")
		return nil
	}
	handler, exists := s.Handlers[eventTyped.GetType()]
	if !exists {
		s.logger.Warn(ctx, "No handler registered for event type %s", eventTyped.GetType())
		return nil
	}
	return handler.Handle(ctx, event, client)
}
