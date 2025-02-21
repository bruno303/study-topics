package hello

import (
	"context"
	"encoding/json"
	"math/rand"
	"net/http"

	"github.com/bruno303/study-topics/go-study/internal/config"
	"github.com/bruno303/study-topics/go-study/internal/crosscutting/observability/log"
	"github.com/bruno303/study-topics/go-study/internal/hello"
	"github.com/bruno303/study-topics/go-study/internal/infra/api"
	"github.com/bruno303/study-topics/go-study/internal/infra/api/middleware"

	"github.com/google/uuid"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func SetupApi(cfg *config.Config, server *http.ServeMux, helloService hello.HelloService) {
	if !cfg.Application.Hello.Api.Enabled {
		log.Log().Info(context.Background(), "Hello api disabled")
		return
	}
	listAllHandler := listAll(helloService)
	createHandler := create(helloService)

	server.Handle("GET /hello", withTrace("GET /hello", buildChain(cfg, listAllHandler)))
	server.Handle("POST /hello", withTrace("POST /hello", buildChain(cfg, createHandler)))
}

func withTrace(pattern string, h http.Handler) http.Handler {
	return otelhttp.WithRouteTag(pattern, otelhttp.NewHandler(h, pattern))
}

func create(helloService hello.HelloService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		result, err := helloService.Hello(r.Context(), uuid.NewString(), rand.Intn(150))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		b, err := json.Marshal(result)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(b)
	})
}

func listAll(helloService hello.HelloService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		result, err := helloService.ListAll(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		response, err := json.Marshal(result)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	})
}

func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

func buildChain(cfg *config.Config, h http.Handler) *api.Chain {
	chain, err := api.NewChain(
		middleware.LogMiddleware(),
		middleware.NewAuthMiddleware(cfg),
		middleware.NewCorrelationIdMiddleware(),
		middleware.NewMiddleware(h),
	)
	panicIfErr(err)
	return chain
}
