package hello

import (
	"fmt"
	"main/internal/config"
	"main/internal/hello"
	"main/internal/infra/api"
	"math/rand"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func HelloHandler() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if strings.ToUpper(r.Method) != "GET" {
			rw.WriteHeader(http.StatusNotFound)
		}
		name := r.URL.Query().Get("name")
		if name == "" {
			name = "World"
		}
		fmt.Fprintf(rw, "Hello, %v", name)
	})
}

func SetupApi(cfg *config.Config, server *http.ServeMux, helloService hello.HelloService, helloRepository hello.Repository) {
	if !cfg.Application.Hello.Api.Enabled {
		return
	}
	pattern := "/hello"

	helloHandler := api.ToMiddlewareFunc(handleHello(helloService, helloRepository))
	helloHandler = api.LogMiddleware(helloHandler)
	helloHandler = api.RequestIdMiddleware(helloHandler)
	chain, err := api.NewChain(helloHandler)
	if err != nil {
		panic(err)
	}
	server.Handle(pattern, withTrace(pattern, chain))
}

func handleHello(helloService hello.HelloService, helloRepository hello.Repository) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch strings.ToUpper(r.Method) {
		case "GET":
			listUsers(helloRepository)(w, r)
		case "POST":
			postUsers(helloService)(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
}

func withTrace(pattern string, h http.Handler) http.Handler {
	return otelhttp.WithRouteTag(pattern, h)
}

func postUsers(helloService hello.HelloService) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		result := helloService.Hello(r.Context(), uuid.NewString(), rand.Intn(150))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(result))
	}
}

func listUsers(helloRepository hello.Repository) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		result := helloRepository.ListAll(r.Context())
		response := ""
		for i, value := range result {
			response += value.ToString()
			if i < len(result)-1 {
				response += ", "
			}
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}
}
