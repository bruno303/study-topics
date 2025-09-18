package bus

import (
	"fmt"
	"sync"
)

type InMemoryHub struct {
	lock  sync.Mutex
	Rooms []*InMemoryRoom
}

func NewHub() *InMemoryHub {
	return &InMemoryHub{
		Rooms: make([]*InMemoryRoom, 0),
	}
}

func (h *InMemoryHub) AddRoom(room *InMemoryRoom) {
	h.lock.Lock()
	defer h.lock.Unlock()
	room.Hub = h
	h.Rooms = append(h.Rooms, room)
}

func (h *InMemoryHub) GetRoom(roomID string) (*InMemoryRoom, error) {
	for _, room := range h.Rooms {
		if room.ID == roomID {
			return room, nil
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
