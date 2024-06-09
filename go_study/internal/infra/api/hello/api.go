package hello

import (
	"context"
	"encoding/json"
	"main/internal/config"
	"main/internal/crosscutting/observability/log"
	"main/internal/hello"
	"main/internal/infra/api"
	"main/internal/infra/api/middleware"
	"math/rand"
	"net/http"

	"github.com/google/uuid"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func SetupApi(cfg *config.Config, server *http.ServeMux, helloService hello.HelloService, helloRepository hello.Repository) {
	if !cfg.Application.Hello.Api.Enabled {
		log.Log().Info(context.Background(), "Hello api disabled")
		return
	}
	listAllHandler := listAll(helloRepository)
	createHandler := create(helloService)

	server.Handle("GET /hello", withTrace("GET /hello", buildChain(listAllHandler)))
	server.Handle("POST /hello", withTrace("POST /hello", buildChain(createHandler)))
}

func withTrace(pattern string, h http.Handler) http.Handler {
	return otelhttp.WithRouteTag(pattern, h)
}

func create(helloService hello.HelloService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		result := helloService.Hello(r.Context(), uuid.NewString(), rand.Intn(150))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(result))
	})
}

func listAll(helloRepository hello.Repository) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		result := helloRepository.ListAll(r.Context())
		response, err := json.Marshal(result)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	})
}

func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

func buildChain(h http.Handler) *api.Chain {
	chain, err := api.NewChain(
		middleware.LogMiddleware(),
		middleware.NewCorrelationIdMiddleware(),
		middleware.NewMiddleware(h),
	)
	panicIfErr(err)
	return chain
}
