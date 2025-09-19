package planningpoker

import (
	"reflect"
	"testing"
)

// Mocks for Room and ClientCollection

type mockClientCollection struct {
	clients []*Client
}

func (m *mockClientCollection) Add(client *Client)     {}
func (m *mockClientCollection) Remove(clientID string) {}
func (m *mockClientCollection) Count() int             { return len(m.clients) }
func (m *mockClientCollection) First() (*Client, bool) {
	if len(m.clients) > 0 {
		return m.clients[0], true
	}
	return nil, false
}
func (m *mockClientCollection) ForEach(f func(client *Client))                      {}
func (m *mockClientCollection) Filter(f func(client *Client) bool) ClientCollection { return m }
func (m *mockClientCollection) Values() []*Client                                   { return m.clients }

func TestNewRoomStateCommand(t *testing.T) {
	vote := "5"
	clients := []*Client{
		{ID: "1", Name: "Alice", Vote: &vote, HasVoted: true, IsSpectator: false, IsOwner: true},
		{ID: "2", Name: "Bob", Vote: nil, HasVoted: false, IsSpectator: true, IsOwner: false},
	}
	owner := clients[0]
	room := &Room{
		ID:           "room1",
		Owner:        "Alice",
		OwnerClient:  owner,
		Clients:      &mockClientCollection{clients: clients},
		CurrentStory: "Story 1",
		Reveal:       true,
	}
	got := NewRoomStateCommand(room)
	want := RoomState{
		Type:         "room-state",
		CurrentStory: "Story 1",
		Reveal:       true,
		Participants: []Participant{
			{ID: "1", Name: "Alice", Vote: &vote, HasVoted: true, IsSpectator: false, IsOwner: true},
			{ID: "2", Name: "Bob", Vote: nil, HasVoted: false, IsSpectator: true, IsOwner: false},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("NewRoomStateCommand() = %+v, want %+v", got, want)
	}
}

func TestNewUpdateNameCommand(t *testing.T) {
	got := NewUpdateNameCommand("Charlie")
	want := UpdateName{
		Type:     "update-name",
		Username: "Charlie",
	}
	if got != want {
		t.Errorf("NewUpdateNameCommand() = %+v, want %+v", got, want)
	}
}

func TestNewUpdateClientIDCommand(t *testing.T) {
	got := NewUpdateClientIDCommand("client-123")
	want := UpdateClientID{
		Type:     "update-client-id",
		ClientID: "client-123",
	}
	if got != want {
		t.Errorf("NewUpdateClientIDCommand() = %+v, want %+v", got, want)
	}
}
