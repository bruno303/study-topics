package redishub_test

import (
	"context"
	"os"
	redishub "planning-poker/internal/infra/boundaries/bus/redis"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func setupRedisClient() *redis.Client {
	addr := os.Getenv("REDIS_HOST")
	if addr == "" {
		addr = "localhost:6379"
	}
	return redis.NewClient(&redis.Options{
		Addr: addr,
		DB:   1,
	})
}

func TestIntegration_RoomLifecycle(t *testing.T) {
	client := setupRedisClient()
	client.FlushDB(context.Background())

	hub, err := redishub.NewRedisHub(context.Background(), client)
	assert.NoError(t, err)

	room := hub.NewRoom(context.Background(), "owner1")
	assert.NotNil(t, room)
	assert.NotEmpty(t, room.ID)

	// Room should be persisted in Redis
	persisted, ok := hub.GetRoom(context.Background(), room.ID)
	assert.True(t, ok)
	assert.Equal(t, room.ID, persisted.ID)

	// Add client and verify mapping
	client1 := room.NewClient("client1")
	room.Clients.Add(client1)
	client1.WithRoom(room)
	hub.AddClient(client1)
	found, ok := hub.FindClientByID("client1")
	assert.True(t, ok)
	assert.Equal(t, client1.ID, found.ID)

	// Remove client and verify cleanup
	hub.RemoveClient(context.Background(), "client1", room.ID)
	_, ok = hub.FindClientByID("client1")
	assert.False(t, ok)

	// Remove room and verify deletion
	hub.RemoveRoom(room.ID)
	_, ok = hub.GetRoom(context.Background(), room.ID)
	assert.False(t, ok)
}

func TestIntegration_RoomTTLExpiry(t *testing.T) {
	client := setupRedisClient()
	client.FlushDB(context.Background())

	hub, err := redishub.NewRedisHub(context.Background(), client)
	assert.NoError(t, err)

	room := hub.NewRoom(context.Background(), "owner2")
	assert.NotNil(t, room)

	// Set short TTL for test
	key := "planning-poker:room:" + room.ID
	client.Expire(context.Background(), key, 2*time.Second)

	time.Sleep(3 * time.Second)
	_, ok := hub.GetRoom(context.Background(), room.ID)
	assert.False(t, ok)
}
