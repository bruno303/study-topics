package entity

import (
	"math"
	"testing"

	"go.uber.org/mock/gomock"
)

// simpleClientCollection is a simple in-memory implementation to avoid import cycles
type simpleClientCollection struct {
	clients []*Client
}

func newSimpleClientCollection(clients ...*Client) *simpleClientCollection {
	return &simpleClientCollection{clients: clients}
}

func (cc *simpleClientCollection) Add(client *Client) {
	cc.clients = append(cc.clients, client)
}

func (cc *simpleClientCollection) Remove(clientID string) {
	for i, c := range cc.clients {
		if c.ID == clientID {
			cc.clients = append(cc.clients[:i], cc.clients[i+1:]...)
			return
		}
	}
}

func (cc *simpleClientCollection) Values() []*Client {
	return cc.clients
}

func (cc *simpleClientCollection) Count() int {
	return len(cc.clients)
}

func (cc *simpleClientCollection) First() (*Client, bool) {
	if len(cc.clients) == 0 {
		return nil, false
	}
	return cc.clients[0], true
}

func (cc *simpleClientCollection) ForEach(f func(client *Client)) {
	for _, client := range cc.clients {
		f(client)
	}
}

func (cc *simpleClientCollection) Filter(f func(client *Client) bool) ClientCollection {
	filtered := newSimpleClientCollection()
	for _, client := range cc.clients {
		if f(client) {
			filtered.Add(client)
		}
	}
	return filtered
}

func TestRoom_RevealWithNoValidVotes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Test case: Single user votes with non-numeric value (NaN scenario)
	cc := newSimpleClientCollection()
	room := NewRoom(cc)
	client := room.NewClient("client1")
	client.Name = "Alice"

	// Vote with a non-numeric value
	vote := "?"
	client.CurrentVote = &vote
	client.HasVoted = true

	// Reveal the votes
	room.reveal(true)

	// The result should be nil when there are no valid numeric votes
	if room.Result != nil {
		if math.IsNaN(float64(*room.Result)) || math.IsInf(float64(*room.Result), 0) {
			t.Errorf("expected Result to be nil when no valid numeric votes, got NaN or Inf: %v", *room.Result)
		} else {
			t.Errorf("expected Result to be nil when no valid numeric votes, got: %v", *room.Result)
		}
	}
}

func TestRoom_RevealWithSingleValidVote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Test case: Single user votes with numeric value
	cc := newSimpleClientCollection()
	room := NewRoom(cc)
	client := room.NewClient("client1")
	client.Name = "Alice"

	// Vote with a numeric value
	vote := "5"
	client.CurrentVote = &vote
	client.HasVoted = true

	// Reveal the votes
	room.reveal(true)

	// The result should be 5
	if room.Result == nil {
		t.Errorf("expected Result to be 5, got nil")
	} else if *room.Result != 5 {
		t.Errorf("expected Result to be 5, got: %v", *room.Result)
	}
}

func TestRoom_RevealWithMixedVotes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Test case: Multiple users with mixed valid and invalid votes
	cc := newSimpleClientCollection()
	room := NewRoom(cc)
	client1 := room.NewClient("client1")
	client1.Name = "Alice"
	client2 := room.NewClient("client2")
	client2.Name = "Bob"

	// One valid vote and one invalid vote
	vote1 := "3"
	vote2 := "coffee"
	client1.CurrentVote = &vote1
	client1.HasVoted = true
	client2.CurrentVote = &vote2
	client2.HasVoted = true

	// Reveal the votes
	room.reveal(true)

	// The result should be 3 (average of only the valid vote)
	if room.Result == nil {
		t.Errorf("expected Result to be 3, got nil")
	} else if *room.Result != 3 {
		t.Errorf("expected Result to be 3, got: %v", *room.Result)
	}
}

func TestRoom_RevealWithAllInvalidVotes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Test case: Multiple users all voting with non-numeric values
	cc := newSimpleClientCollection()
	room := NewRoom(cc)
	client1 := room.NewClient("client1")
	client1.Name = "Alice"
	client2 := room.NewClient("client2")
	client2.Name = "Bob"

	// Both vote with non-numeric values
	vote1 := "?"
	vote2 := "coffee"
	client1.CurrentVote = &vote1
	client1.HasVoted = true
	client2.CurrentVote = &vote2
	client2.HasVoted = true

	// Reveal the votes
	room.reveal(true)

	// The result should be nil when there are no valid numeric votes
	if room.Result != nil {
		if math.IsNaN(float64(*room.Result)) || math.IsInf(float64(*room.Result), 0) {
			t.Errorf("expected Result to be nil when no valid numeric votes, got NaN or Inf: %v", *room.Result)
		} else {
			t.Errorf("expected Result to be nil when no valid numeric votes, got: %v", *room.Result)
		}
	}
}

func TestRoom_RevealWithSpectators(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Test case: Spectators should not be counted in results
	cc := newSimpleClientCollection()
	room := NewRoom(cc)
	client1 := room.NewClient("client1")
	client1.Name = "Alice"
	client2 := room.NewClient("client2")
	client2.Name = "Bob"
	client2.IsSpectator = true

	// Alice votes with numeric value, Bob (spectator) votes with numeric value
	vote1 := "5"
	vote2 := "10"
	client1.CurrentVote = &vote1
	client1.HasVoted = true
	client2.CurrentVote = &vote2
	client2.HasVoted = true

	// Reveal the votes
	room.reveal(true)

	// The result should be 5 (only Alice's vote counts)
	if room.Result == nil {
		t.Errorf("expected Result to be 5, got nil")
	} else if *room.Result != 5 {
		t.Errorf("expected Result to be 5, got: %v", *room.Result)
	}
}
