package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/bruno303/go-toolkit/pkg/log"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	"planning-poker/internal/application/planningpoker/usecase"
	"planning-poker/internal/domain"
	"planning-poker/internal/infra/bus"
)

type (
	contextKey string

	websocketBusFactory interface {
		NewBus(input bus.WebSocketBusFactoryInput) domain.Bus
	}

	WebsocketAPI struct {
		upgrader   websocket.Upgrader
		usecases   usecase.UseCasesFacade
		busFactory websocketBusFactory
		logger     log.Logger
	}
)

var _ API = (*WebsocketAPI)(nil)

func NewWebsocketAPI(usecases usecase.UseCasesFacade, websocketBusFactory websocketBusFactory) *WebsocketAPI {
	return &WebsocketAPI{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		usecases:   usecases,
		busFactory: websocketBusFactory,
		logger:     log.NewLogger("planningpoker.api.websocket"),
	}
}

func (api *WebsocketAPI) Endpoint() string {
	return "/planning/{roomID}/ws"
}

func (api *WebsocketAPI) Methods() []string {
	return nil
}

func (api *WebsocketAPI) Handle() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		roomID := mux.Vars(r)["roomID"]
		if roomID == "" {
			SendJsonErrorMsg(w, http.StatusBadRequest, "Room ID is required")
			return
		}

		handleWebsocketConnection(api.logger, api.upgrader, api.usecases, api.busFactory, w, r, roomID)
	})
}

func handleWebsocketConnection(logger log.Logger, upgrader websocket.Upgrader, usecases usecase.UseCasesFacade, busFactory websocketBusFactory, w http.ResponseWriter, r *http.Request, roomID string) {
	r = r.WithContext(context.WithValue(r.Context(), contextKey("roomID"), roomID))
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error(r.Context(), "Error upgrading to WebSocket", err)
		return
	}

	createClientOutput, err := usecases.CreateClient.Execute(r.Context())
	if err != nil {
		logger.Error(r.Context(), "Error creating client", err)
		SendJsonErrorMsg(w, http.StatusInternalServerError, "Error creating client")
		return
	}

	wsBus := busFactory.NewBus(bus.WebSocketBusFactoryInput{
		ClientID: createClientOutput.ClientID,
		RoomID:   roomID,
		Socket:   ws,
	})

	output, err := usecases.JoinRoom.Execute(r.Context(), usecase.JoinRoomCommand{
		RoomID:   roomID,
		SenderID: createClientOutput.ClientID,
		Bus:      wsBus,
	})
	if err != nil {
		logger.Error(r.Context(), fmt.Sprintf("Error joining room %s", roomID), err)
		closeMsg := websocket.FormatCloseMessage(websocket.CloseInternalServerErr, fmt.Sprintf("Error joining room %s", roomID))
		_ = ws.WriteControl(websocket.CloseMessage, closeMsg, time.Now().Add(time.Second))
		_ = ws.Close()
		return
	}

	logger.Info(r.Context(), "New client connected: %v on room: %v", output.Client.ID, output.Room.ID)

	wsBus.Listen(r.Context())
}
