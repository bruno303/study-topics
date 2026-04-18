package http

import (
	"net/http"

	"github.com/bruno303/go-toolkit/pkg/log"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"planning-poker/internal/application/planningpoker/usecase"
)

type WebsocketJoinAPI struct {
	upgrader   websocket.Upgrader
	usecases   usecase.UseCasesFacade
	busFactory websocketBusFactory
	logger     log.Logger
}

var _ API = (*WebsocketJoinAPI)(nil)

func NewWebsocketJoinAPI(usecases usecase.UseCasesFacade, websocketBusFactory websocketBusFactory) *WebsocketJoinAPI {
	return &WebsocketJoinAPI{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		usecases:   usecases,
		busFactory: websocketBusFactory,
		logger:     log.NewLogger("planningpoker.api.websocketjoin"),
	}
}

func (api *WebsocketJoinAPI) Endpoint() string {
	return "/planning/join"
}

func (api *WebsocketJoinAPI) Methods() []string {
	return nil
}

func (api *WebsocketJoinAPI) Handle() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleWebsocketConnection(api.logger, api.upgrader, api.usecases, api.busFactory, w, r, uuid.NewString())
	})
}
