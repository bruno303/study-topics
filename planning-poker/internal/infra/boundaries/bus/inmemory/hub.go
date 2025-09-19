package inmemory

import (
	"fmt"
	"planning-poker/internal/application/planningpoker"
	"sync"
)

type InMemoryHub struct {
	lock  sync.Mutex
	Rooms []*planningpoker.Room
}

func NewHub() *InMemoryHub {
	return &InMemoryHub{
		Rooms: make([]*planningpoker.Room, 0),
	}
}

func (h *InMemoryHub) NewRoom(owner string) *planningpoker.Room {
	h.lock.Lock()
	defer h.lock.Unlock()

	room := planningpoker.NewRoom(owner, NewInMemoryClientCollection())
	room.Hub = h
	h.Rooms = append(h.Rooms, room)
	return room
}

func (h *InMemoryHub) GetRoom(roomID string) (*planningpoker.Room, error) {
	for _, room := range h.Rooms {
		if room.ID == roomID {
			return planningpoker.NewRoom(room.Owner, NewInMemoryClientCollection(room.Clients.Values()...)), nil
		}
	}
	return nil, fmt.Errorf("room %s not found", roomID)
}

func (h *InMemoryHub) RemoveRoom(roomID string) {
	h.lock.Lock()
	defer h.lock.Unlock()

	for i, room := range h.Rooms {
		if room.ID == roomID {
			h.Rooms = append(h.Rooms[:i], h.Rooms[i+1:]...)
			return
		}
	}
}
