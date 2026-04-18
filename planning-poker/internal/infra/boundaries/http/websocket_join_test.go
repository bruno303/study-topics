package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"go.uber.org/mock/gomock"

	"planning-poker/internal/application/planningpoker/usecase"
	"planning-poker/internal/domain"
	"planning-poker/internal/domain/entity"
	"planning-poker/internal/infra/bus"
)

type testWebsocketBusFactory struct {
	newBus func(input bus.WebSocketBusFactoryInput) domain.Bus
}

func (f testWebsocketBusFactory) NewBus(input bus.WebSocketBusFactoryInput) domain.Bus {
	return f.newBus(input)
}

type noopBus struct {
	roomID string
}

func (b noopBus) Close() error { return nil }

func (b noopBus) Send(context.Context, any) error { return nil }

func (b noopBus) Listen(context.Context) {}

func (b noopBus) RoomID() string { return b.roomID }

func TestWebsocketJoinAPI_Endpoint(t *testing.T) {
	api := NewWebsocketJoinAPI(usecase.UseCasesFacade{}, testWebsocketBusFactory{})

	if got, want := api.Endpoint(), "/planning/join"; got != want {
		t.Fatalf("Endpoint() = %v, want %v", got, want)
	}
}

func TestWebsocketJoinAPI_Methods(t *testing.T) {
	api := NewWebsocketJoinAPI(usecase.UseCasesFacade{}, testWebsocketBusFactory{})

	if methods := api.Methods(); methods != nil {
		t.Fatalf("Methods() = %#v, want nil", methods)
	}
}

func TestWebsocketJoinAPI_Handle_GeneratesRoomIDAndJoins(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCreateClient := usecase.NewMockUseCaseO[usecase.CreateClientOutput](ctrl)
	mockJoinRoom := usecase.NewMockUseCaseR[usecase.JoinRoomCommand, *usecase.JoinRoomOutput](ctrl)

	createClientOutput := usecase.CreateClientOutput{ClientID: "client-123"}
	mockCreateClient.EXPECT().Execute(gomock.Any()).Return(createClientOutput, nil)

	var capturedInput bus.WebSocketBusFactoryInput
	var capturedCmd usecase.JoinRoomCommand
	joined := make(chan struct{})
	factory := testWebsocketBusFactory{newBus: func(input bus.WebSocketBusFactoryInput) domain.Bus {
		capturedInput = input
		return noopBus{roomID: input.RoomID}
	}}

	mockJoinRoom.EXPECT().Execute(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, cmd usecase.JoinRoomCommand) (*usecase.JoinRoomOutput, error) {
		capturedCmd = cmd
		close(joined)
		return &usecase.JoinRoomOutput{
			Client: &entity.Client{ID: cmd.SenderID},
			Room:   &entity.Room{ID: cmd.RoomID},
		}, nil
	})

	api := NewWebsocketJoinAPI(usecase.UseCasesFacade{
		CreateClient: mockCreateClient,
		JoinRoom:     mockJoinRoom,
	}, factory)

	router := mux.NewRouter()
	router.Handle(api.Endpoint(), api.Handle())
	ts := httptest.NewServer(router)
	defer ts.Close()

	wsURL := strings.Replace(ts.URL, "http://", "ws://", 1)
	conn, _, err := websocket.DefaultDialer.Dial(wsURL+"/planning/join", nil)
	if err != nil {
		t.Fatalf("failed to connect to websocket: %v", err)
	}
	defer func() { _ = conn.Close() }()

	select {
	case <-joined:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for join flow to complete")
	}

	if capturedCmd.RoomID == "" {
		t.Fatal("expected generated roomID to be set")
	}
	if _, err := uuid.Parse(capturedCmd.RoomID); err != nil {
		t.Fatalf("expected generated roomID to be a UUID, got %q: %v", capturedCmd.RoomID, err)
	}
	if capturedCmd.RoomID != capturedInput.RoomID {
		t.Fatalf("expected bus factory roomID %q, got %q", capturedCmd.RoomID, capturedInput.RoomID)
	}
	if capturedCmd.SenderID != createClientOutput.ClientID {
		t.Fatalf("expected senderID %q, got %q", createClientOutput.ClientID, capturedCmd.SenderID)
	}
	if capturedInput.ClientID != createClientOutput.ClientID {
		t.Fatalf("expected bus factory clientID %q, got %q", createClientOutput.ClientID, capturedInput.ClientID)
	}
	if capturedCmd.Bus == nil {
		t.Fatal("expected join command bus to be set")
	}
	if got := capturedCmd.Bus.RoomID(); got != capturedCmd.RoomID {
		t.Fatalf("expected bus roomID %q, got %q", capturedCmd.RoomID, got)
	}
}

func TestWebsocketJoinAPI_Handle_InvalidUpgrade(t *testing.T) {
	api := NewWebsocketJoinAPI(usecase.UseCasesFacade{}, testWebsocketBusFactory{})
	router := mux.NewRouter()
	router.Handle(api.Endpoint(), api.Handle())
	ts := httptest.NewServer(router)
	defer ts.Close()

	resp, err := http.Get(ts.URL + api.Endpoint())
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status code = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestWebsocketJoinAPI_Handle_JoinFailureClosesConnection(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCreateClient := usecase.NewMockUseCaseO[usecase.CreateClientOutput](ctrl)
	mockJoinRoom := usecase.NewMockUseCaseR[usecase.JoinRoomCommand, *usecase.JoinRoomOutput](ctrl)

	mockCreateClient.EXPECT().Execute(gomock.Any()).Return(usecase.CreateClientOutput{ClientID: "client-123"}, nil)
	mockJoinRoom.EXPECT().Execute(gomock.Any(), gomock.Any()).Return((*usecase.JoinRoomOutput)(nil), errors.New("join failed"))

	api := NewWebsocketJoinAPI(usecase.UseCasesFacade{
		CreateClient: mockCreateClient,
		JoinRoom:     mockJoinRoom,
	}, testWebsocketBusFactory{newBus: func(input bus.WebSocketBusFactoryInput) domain.Bus {
		return noopBus{roomID: input.RoomID}
	}})

	router := mux.NewRouter()
	router.Handle(api.Endpoint(), api.Handle())
	ts := httptest.NewServer(router)
	defer ts.Close()

	wsURL := strings.Replace(ts.URL, "http://", "ws://", 1)
	conn, _, err := websocket.DefaultDialer.Dial(wsURL+api.Endpoint(), nil)
	if err != nil {
		t.Fatalf("failed to connect to websocket: %v", err)
	}
	defer func() { _ = conn.Close() }()

	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, _, readErr := conn.ReadMessage()
	if readErr == nil {
		t.Fatal("expected websocket close error, got nil")
	}
	if !websocket.IsCloseError(readErr, websocket.CloseInternalServerErr) {
		t.Fatalf("expected close error %d, got %v", websocket.CloseInternalServerErr, readErr)
	}
}
