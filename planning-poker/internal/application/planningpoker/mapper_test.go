package planningpoker

import (
	"reflect"
	"testing"

	"github.com/samber/lo"
)

func TestMapToParticipants(t *testing.T) {
	clients := []*Client{
		{ID: "1", Name: "Alice", Vote: lo.ToPtr("5"), HasVoted: true, IsSpectator: false, IsOwner: false},
		{ID: "2", Name: "Bob", Vote: lo.ToPtr("3"), HasVoted: true, IsSpectator: false, IsOwner: true},
		{ID: "3", Name: "Charlie", Vote: lo.ToPtr("?"), HasVoted: false, IsSpectator: true, IsOwner: false},
	}
	owner := clients[1]

	participants := MapToParticipants(clients, owner)

	expected := []Participant{
		{ID: "1", Name: "Alice", Vote: lo.ToPtr("5"), HasVoted: true, IsSpectator: false, IsOwner: false},
		{ID: "2", Name: "Bob", Vote: lo.ToPtr("3"), HasVoted: true, IsSpectator: false, IsOwner: true},
		{ID: "3", Name: "Charlie", Vote: lo.ToPtr("?"), HasVoted: false, IsSpectator: true, IsOwner: false},
	}

	if !reflect.DeepEqual(participants, expected) {
		t.Errorf("Expected %v, got %v", expected, participants)
	}
}
