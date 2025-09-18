package http

import (
	"context"
	"encoding/json"
	"net/http"
	"planning-poker/internal/infra/boundaries/bus"
	"planning-poker/internal/planningpoker"

	"github.com/bruno303/go-toolkit/pkg/log"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	logger = log.NewLogger("planningpoker.api")
)

type (
	contextKey string

	hub interface {
		NewRoom(owner string) *planningpoker.Room
		GetRoom(roomID string) (*planningpoker.Room, error)
		RemoveRoom(roomID string)
	}
)

func ConfigurePlanningPokerAPI(mux *mux.Router, hub hub) {
	mux.HandleFunc("/planning/{roomID}/ws", handleConnections(hub))
	mux.HandleFunc("/planning/room", createRoom(hub)).Methods("POST", "OPTIONS")
	mux.HandleFunc("/planning/room/{roomID}", getRoom(hub)).Methods("GET")
}

func handleConnections(hub hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		roomID := mux.Vars(r)["roomID"]
		if roomID == "" {
			http.Error(w, "Room ID is required", http.StatusBadRequest)
			return
		}

		r = r.WithContext(context.WithValue(r.Context(), contextKey("roomID"), roomID))
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logger.Error(r.Context(), "Error upgrading to WebSocket", err)
		}

		room, err := hub.GetRoom(roomID)
		if err != nil {
			logger.Error(r.Context(), "Error getting room", err)
			return
		}

		bus := bus.NewWebsocketBus(ws)
		client := room.NewClient(bus)
		logger.Info(r.Context(), "New client connected: %v on room: %v", client.ID, room.ID)

		client.Listen(r.Context())
	}
}

type CreateRoomRequest struct {
	CreatedBy string `json:"createdBy"`
}

type CreateRoomResponse struct {
	RoomID string `json:"roomId"`
}

func createRoom(hub hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var body CreateRoomRequest
		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		room := hub.NewRoom(body.CreatedBy)
		logger.Info(r.Context(), "New room created: %v", room.ID)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(CreateRoomResponse{RoomID: room.ID})
	}
}

func getRoom(hub hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		roomID := mux.Vars(r)["roomID"]
		if roomID == "" {
			http.Error(w, "Room ID is required", http.StatusBadRequest)
			return
		}

		_, err := hub.GetRoom(roomID)
		if err != nil {
			http.Error(w, "Room not found", http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
