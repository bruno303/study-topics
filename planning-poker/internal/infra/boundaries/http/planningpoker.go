package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"planning-poker/internal/application/planningpoker/shared"
	"planning-poker/internal/application/planningpoker/usecase"
	"planning-poker/internal/application/planningpoker/usecase/dto"

	"github.com/bruno303/go-toolkit/pkg/log"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/samber/lo"
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

	WebsocketBus struct {
		ID       string
		conn     *websocket.Conn
		hub      shared.Hub
		usecases usecase.UseCases
		logger   log.Logger
	}
)

func ConfigurePlanningPokerAPI(mux *mux.Router, hub shared.Hub, usecases usecase.UseCases) {
	mux.HandleFunc("/planning/{roomID}/ws", handleConnections(hub, usecases))
	mux.HandleFunc("/planning/room", createRoom(hub)).Methods("POST", "OPTIONS")
	mux.HandleFunc("/planning/room/{roomID}", getRoom(hub)).Methods("GET")
}

func handleConnections(hub shared.Hub, usecases usecase.UseCases) http.HandlerFunc {
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

		room, ok := hub.GetRoom(roomID)
		if !ok {
			logger.Error(r.Context(), "Error getting room", err)
			return
		}

		clientID := uuid.NewString()
		client := room.NewClient(clientID)
		hub.AddClient(client)

		bus := NewWebsocketBus(client.ID, ws, hub, usecases)
		if err != nil {
			logger.Error(r.Context(), fmt.Sprintf("Error creating bus for client %s", clientID), err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "{ \"msg\": \"Error creating bus for client %s\" }", clientID)
			return
		}

		hub.AddBus(client.ID, bus)
		logger.Info(r.Context(), "New client connected: %v on room: %v", client.ID, room.ID)

		defer bus.Close()
		defer func() {
			hub.RemoveClient(r.Context(), client.ID, room.ID)
		}()

		bus.Send(r.Context(), dto.NewUpdateClientIDCommand(client.ID))
		bus.Listen(r.Context())
	}
}

type CreateRoomRequest struct {
	CreatedBy string `json:"createdBy"`
}

type CreateRoomResponse struct {
	RoomID string `json:"roomId"`
}

func createRoom(hub shared.Hub) http.HandlerFunc {
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

func getRoom(hub shared.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		roomID := mux.Vars(r)["roomID"]
		if roomID == "" {
			http.Error(w, "Room ID is required", http.StatusBadRequest)
			return
		}

		_, ok := hub.GetRoom(roomID)
		if !ok {
			http.Error(w, "Room not found", http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func NewWebsocketBus(
	id string,
	socket *websocket.Conn,
	hub shared.Hub,
	usecases usecase.UseCases,
) *WebsocketBus {
	return &WebsocketBus{
		ID:       id,
		conn:     socket,
		hub:      hub,
		usecases: usecases,
		logger:   log.NewLogger("websocket.client"),
	}
}

func (c *WebsocketBus) Close() error {
	return c.conn.Close()
}

func (c *WebsocketBus) Send(ctx context.Context, message any) error {
	c.logger.Debug(ctx, "Sending message to client: %v", message)
	return c.conn.WriteJSON(message)
}

func (c *WebsocketBus) receive() (map[string]any, error) {
	var msg map[string]any
	err := c.conn.ReadJSON(&msg)
	return msg, err
}

func (c *WebsocketBus) Listen(ctx context.Context) {
	defer c.Close()

	for {
		msg, err := c.receive()
		if err != nil {
			if _, ok := err.(*websocket.CloseError); ok {
				c.logger.Info(ctx, "Client %v disconnected", c)

			} else {
				c.logger.Error(ctx, fmt.Sprintf("Error receiving message from client %v", c.ID), err)
			}

			return
		}
		c.logger.Info(ctx, "Message received from client %v: %v", c, msg)

		eventType, ok := msg["type"].(string)
		if !ok {
			c.logger.Error(ctx, fmt.Sprintf("Error casting message type to string for client %v", c.ID), err)
			continue
		}

		var uerr error
		switch eventType {
		case "update-name":
			uerr = c.usecases.UpdateName.Execute(ctx, usecase.UpdateNameCommand{
				ClientID: msg["clientId"].(string),
				Username: msg["username"].(string),
			})
		case "vote":
			uerr = c.usecases.Vote.Execute(ctx, usecase.VoteCommand{
				ClientID: msg["clientId"].(string),
				Vote:     lo.ToPtr(msg["vote"].(string)),
			})
		case "reset":
			uerr = c.usecases.Reset.Execute(ctx, usecase.ResetCommand{
				ClientID: msg["clientId"].(string),
			})
		case "reveal-votes":
			uerr = c.usecases.Reveal.Execute(ctx, usecase.RevealCommand{
				ClientID: msg["clientId"].(string),
			})
		case "toggle-spectator":
			uerr = c.usecases.ToggleSpectator.Execute(ctx, usecase.ToggleSpectatorCommand{
				ClientID:       msg["clientId"].(string),
				TargetClientID: msg["targetClientId"].(string),
			})
		case "toggle-owner":
			uerr = c.usecases.ToggleOwner.Execute(ctx, usecase.ToggleOwnerCommand{
				ClientID:       msg["clientId"].(string),
				TargetClientID: msg["targetClientId"].(string),
			})
		case "update-story":
			uerr = c.usecases.UpdateStory.Execute(ctx, usecase.UpdateStoryCommand{
				ClientID: msg["clientId"].(string),
				Story:    msg["story"].(string),
			})
		case "new-voting":
			uerr = c.usecases.NewVoting.Execute(ctx, usecase.NewVotingCommand{
				ClientID: msg["clientId"].(string),
			})
		case "vote-again":
			uerr = c.usecases.VoteAgain.Execute(ctx, usecase.VoteAgainCommand{
				ClientID: msg["clientId"].(string),
			})
		default:
			c.logger.Error(ctx, fmt.Sprintf("Unknown event type '%v' for client %v", eventType, c.ID), errors.New("unknown event type"))
			continue
		}

		if uerr != nil {
			c.logger.Error(ctx, fmt.Sprintf("Error handling event for client %v", c.ID), uerr)
			continue
		}
	}
}
