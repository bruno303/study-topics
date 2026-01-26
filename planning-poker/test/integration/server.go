package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"planning-poker/internal/config"
	"planning-poker/internal/infra/boundaries/bus/redis"
	"planning-poker/internal/setup"
	"testing"

	"github.com/gorilla/mux"
)

type TestServer struct {
	Server    *httptest.Server
	Router    *mux.Router
	Container *setup.Container
}

func NewTestServer(t *testing.T) *TestServer {
	t.Helper()
	cfg := getTestConfig()
	setup.ConfigureLogging(cfg)

	cleanRedis(t, cfg)

	container := setup.NewContainer(cfg)

	r := mux.NewRouter()
	setup.ConfigureAPIs(r, container)
	ts := httptest.NewServer(r)

	return &TestServer{
		Server:    ts,
		Router:    r,
		Container: container,
	}
}

func (ts *TestServer) Close() {
	ts.Server.Close()
	if hub, ok := ts.Container.Hub.(*redis.RedisHub); ok {
		_ = hub.Close()
	}
}

func cleanRedis(t *testing.T, cfg *config.Config) {
	t.Helper()

	redisClient, err := setup.NewRedisClient(cfg)
	if err != nil {
		t.Fatalf("failed to create redis client for cleanup: %v", err)
	}
	defer func() { _ = redisClient.Close() }()

	if err := redisClient.FlushDB(context.Background()).Err(); err != nil {
		t.Fatalf("failed to flush redis: %v", err)
	}
}

func (ts *TestServer) GetJSON(t *testing.T, path string, target any) (*http.Response, error) {
	t.Helper()

	resp, err := http.Get(ts.Server.URL + path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return resp, fmt.Errorf("failed to decode JSON: %w", err)
	}

	return resp, nil
}

func getTestConfig() *config.Config {
	err := os.Setenv("CONFIG_FILE", "config-test.yaml")
	if err != nil {
		panic(fmt.Sprintf("failed to set CONFIG_FILE env: %v", err))
	}
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(fmt.Sprintf("failed to load test config: %v", err))
	}
	return cfg
}
