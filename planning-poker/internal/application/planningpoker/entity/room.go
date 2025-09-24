package entity

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/samber/lo"
)

type (
	ClientCollection interface {
		Add(client *Client)
		Remove(clientID string)
		Count() int
		First() (*Client, bool)
		ForEach(f func(client *Client))
		Filter(f func(client *Client) bool) ClientCollection
		Values() []*Client
	}

	Room struct {
		ID           string
		Owner        string
		OwnerClient  *Client
		Clients      ClientCollection
		CurrentStory string
		Reveal       bool
		lock         sync.Mutex
	}
)

func NewRoom(owner string, clients ClientCollection) *Room {
	return &Room{
		ID:           uuid.NewString(),
		Owner:        owner,
		Clients:      clients,
		CurrentStory: "",
		Reveal:       false,
	}
}

func (r *Room) NewClient(id string) *Client {
	r.lock.Lock()
	defer r.lock.Unlock()

	client := newClient(id)
	r.Clients.Add(client)
	client.room = r

	if r.Clients.Count() == 1 {
		client.IsOwner = true
	}

	return client
}

func (r *Room) RemoveClient(ctx context.Context, clientID string) {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.Clients.Remove(clientID)

	if r.CountOwners() == 0 {
		if client, ok := r.Clients.First(); ok {
			client.IsOwner = true
		}
	}
}

func (r *Room) NewVoting() {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.CurrentStory = ""
	r.Reveal = false
	r.Clients.ForEach(func(c *Client) {
		c.HasVoted = false
		c.vote = nil
	})
}

func (r *Room) ResetVoting() {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.Reveal = false

	r.Clients.ForEach(func(c *Client) {
		c.HasVoted = false
		c.vote = nil
	})
}

func (r *Room) checkReveal() {
	r.lock.Lock()
	defer r.lock.Unlock()

	activeClients := r.Clients.Filter(func(client *Client) bool {
		return !client.IsSpectator
	})

	if lo.EveryBy(activeClients.Values(), func(client *Client) bool {
		return client.HasVoted
	}) {
		r.Reveal = true
	}
}

func (r *Room) CountOwners() int {
	return r.Clients.Filter(func(client *Client) bool {
		return client.IsOwner
	}).Count()
}

func (r *Room) ToggleSpectator(ctx context.Context, clientID string) {
	r.Clients.Filter(func(client *Client) bool {
		return client.ID == clientID
	}).ForEach(func(client *Client) {
		client.IsSpectator = !client.IsSpectator
		client.Vote(ctx, nil)
	})
}

func (r *Room) ToggleOwner(clientID string) {
	owners := r.Clients.Filter(func(client *Client) bool {
		return client.IsOwner
	})

	ownerCount := owners.Count()

	if ownerCount == 1 {
		if first, ok := owners.First(); ok && first.ID == clientID && first.IsOwner {
			// Prevent removing the last owner
			return
		}
	}

	owners.ForEach(func(client *Client) {
		client.IsOwner = !client.IsOwner
	})
}

func (r *Room) SetCurrentStory(story string) {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.CurrentStory = story
}

func (r *Room) ToggleReveal() {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.Reveal = !r.Reveal
}

func (r *Room) IsEmpty() bool {
	return r.Clients.Count() == 0
}
