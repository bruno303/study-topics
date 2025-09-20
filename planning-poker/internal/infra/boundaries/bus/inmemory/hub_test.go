package inmemory

import (
	"testing"
)

func TestNewRoom(t *testing.T) {
	hub := NewHub()
	owner := "test-owner"

	room := hub.NewRoom(owner)

	if room == nil {
		t.Fatal("expected room to be non-nil")
	}
	if room.Owner != owner {
		t.Errorf("expected room owner to be %q, got %q", owner, room.Owner)
	}
	if room.Hub != hub {
		t.Errorf("expected room.Hub to be the hub instance")
	}
	if len(hub.Rooms) != 1 {
		t.Errorf("expected hub.Rooms to have 1 room, got %d", len(hub.Rooms))
	}
	if hub.Rooms[0] != room {
		t.Errorf("expected hub.Rooms[0] to be the created room")
	}
	if room.Clients == nil {
		t.Error("expected room.Clients to be initialized")
	}
}

func TestGetRoom(t *testing.T) {
	hub := NewHub()
	owner := "owner1"
	room := hub.NewRoom(owner)

	got, err := hub.GetRoom(room.ID)
	if err != nil {
		t.Fatalf("expected to find room, got error: %v", err)
	}
	if got != room {
		t.Errorf("expected to get the created room, got different room")
	}

	_, err = hub.GetRoom("non-existent-id")
	if err == nil {
		t.Error("expected error for non-existent room ID, got nil")
	}
}

func TestRemoveRoom(t *testing.T) {
	hub := NewHub()
	room1 := hub.NewRoom("owner1")
	room2 := hub.NewRoom("owner2")

	if len(hub.Rooms) != 2 {
		t.Fatalf("expected 2 rooms, got %d", len(hub.Rooms))
	}

	hub.RemoveRoom(room1.ID)
	if len(hub.Rooms) != 1 {
		t.Errorf("expected 1 room after removal, got %d", len(hub.Rooms))
	}
	if hub.Rooms[0] != room2 {
		t.Errorf("expected remaining room to be room2")
	}

	// Remove non-existent room should not panic or change state
	hub.RemoveRoom("non-existent-id")
	if len(hub.Rooms) != 1 {
		t.Errorf("expected 1 room after removing non-existent, got %d", len(hub.Rooms))
	}
}
