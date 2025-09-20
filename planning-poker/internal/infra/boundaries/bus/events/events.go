package events

type (
	Event interface {
		GetType() string
	}

	EventTypeAware struct {
		Type string `json:"type"`
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

// func (e InitEvent) Type() string      { return "init" }
// func (e VoteEvent) Type() string      { return "vote" }
// func (e RevealEvent) Type() string    { return "reveal-votes" }
// func (e ResetEvent) Type() string     { return "reset" }
// func (e SpectatorEvent) Type() string { return "toggle-spectator" }
// func (e StoryEvent) Type() string     { return "update-story" }
// func (e OwnerEvent) Type() string     { return "toggle-owner" }
// func (e NewVotingEvent) Type() string { return "new-voting" }
// func (e VoteAgainEvent) Type() string { return "vote-again" }
