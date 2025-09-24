package dto

import (
	"context"
	"planning-poker/internal/application/planningpoker/entity"
	"reflect"
	"testing"

	"github.com/samber/lo"
)

// Mocks for Room and ClientCollection

type mockClientCollection struct {
	clients []*entity.Client
}

func (m *mockClientCollection) Add(client *entity.Client) {}
func (m *mockClientCollection) Remove(clientID string)    {}
func (m *mockClientCollection) Count() int                { return len(m.clients) }
func (m *mockClientCollection) First() (*entity.Client, bool) {
	if len(m.clients) > 0 {
		return m.clients[0], true
	}
	return nil, false
}
func (m *mockClientCollection) ForEach(f func(client *entity.Client)) {}
func (m *mockClientCollection) Filter(f func(client *entity.Client) bool) entity.ClientCollection {
	return m
}
func (m *mockClientCollection) Values() []*entity.Client { return m.clients }

func TestNewRoomStateCommand(t *testing.T) {
	vote := "5"
	clients := []*entity.Client{
		{ID: "1", Name: "Alice", IsSpectator: false, IsOwner: true},
		{ID: "2", Name: "Bob", IsSpectator: true, IsOwner: false},
	}

	clients[0].Vote(context.Background(), &vote)
	clients[1].Vote(context.Background(), nil)

	owner := clients[0]
	room := &entity.Room{
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

func TestMapToParticipants(t *testing.T) {
	ctx := context.Background()

	clients := []*entity.Client{
		{ID: "1", Name: "Alice", IsSpectator: false, IsOwner: false},
		{ID: "2", Name: "Bob", IsSpectator: false, IsOwner: true},
		{ID: "3", Name: "Charlie", IsSpectator: true, IsOwner: false},
	}
	clients[0].Vote(ctx, lo.ToPtr("5"))
	clients[1].Vote(ctx, lo.ToPtr("3"))
	clients[2].Vote(ctx, lo.ToPtr("?"))

	owner := clients[1]

	participants := MapToParticipants(clients, owner)

	expected := []Participant{
		{ID: "1", Name: "Alice", Vote: lo.ToPtr("5"), HasVoted: true, IsSpectator: false, IsOwner: false},
		{ID: "2", Name: "Bob", Vote: lo.ToPtr("3"), HasVoted: true, IsSpectator: false, IsOwner: true},
		{ID: "3", Name: "Charlie", Vote: lo.ToPtr("?"), HasVoted: true, IsSpectator: true, IsOwner: false},
	}

	if !reflect.DeepEqual(participants, expected) {
		t.Errorf("Expected %v, got %v", expected, participants)
	}
}
