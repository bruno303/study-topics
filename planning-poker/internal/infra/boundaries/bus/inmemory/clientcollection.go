package inmemory

import "planning-poker/internal/application/planningpoker"

type InMemoryClientCollection struct {
	clients []*planningpoker.Client
}

func NewInMemoryClientCollection(clients ...*planningpoker.Client) *InMemoryClientCollection {
	return &InMemoryClientCollection{
		clients: clients,
	}
}

func (cc *InMemoryClientCollection) Add(client *planningpoker.Client) {
	cc.clients = append(cc.clients, client)
}

func (cc *InMemoryClientCollection) Remove(clientID string) {
	for i, c := range cc.clients {
		if c.ID == clientID {
			cc.clients = append(cc.clients[:i], cc.clients[i+1:]...)
			return
		}
	}
}

func (cc *InMemoryClientCollection) Values() []*planningpoker.Client {
	return cc.clients
}

func (cc *InMemoryClientCollection) Count() int {
	return len(cc.clients)
}

func (cc *InMemoryClientCollection) First() (*planningpoker.Client, bool) {
	if len(cc.clients) == 0 {
		return nil, false
	}
	return cc.clients[0], true
}

func (cc *InMemoryClientCollection) ForEach(f func(client *planningpoker.Client)) {
	for _, client := range cc.clients {
		f(client)
	}
}

func (cc *InMemoryClientCollection) Filter(f func(client *planningpoker.Client) bool) planningpoker.ClientCollection {
	filtered := NewInMemoryClientCollection()
	for _, client := range cc.clients {
		if f(client) {
			filtered.Add(client)
		}
	}
	return filtered
}
