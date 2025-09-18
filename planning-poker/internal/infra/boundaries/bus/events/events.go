package events

type (
	Event interface {
		Type() string
	}

	InitEvent struct {
		Payload struct {
			Username string `json:"username"`
		}
	}

	VoteEvent struct {
		Payload struct {
			Vote string `json:"vote"`
		}
	}

	RevealEvent struct{}

	ResetEvent struct{}

	SpectatorEvent struct {
		Payload struct {
			ClientID string `json:"clientId"`
		}
	}

	StoryEvent struct {
		Payload struct {
			Story string `json:"story"`
		}
	}

	OwnerEvent struct {
		Payload struct {
			ClientID string `json:"clientId"`
		}
	}
)

func (e InitEvent) Type() string      { return "init" }
func (e VoteEvent) Type() string      { return "vote" }
func (e RevealEvent) Type() string    { return "reveal" }
func (e ResetEvent) Type() string     { return "reset" }
func (e SpectatorEvent) Type() string { return "toggle-spectator" }
func (e StoryEvent) Type() string     { return "story" }
func (e OwnerEvent) Type() string     { return "toggle-owner" }
