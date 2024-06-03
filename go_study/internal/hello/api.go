package hello

import (
	"fmt"
	"main/internal/config"
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

func SetupApi(cfg *config.Config, server *http.ServeMux, helloService HelloService, helloRepository Repository) {
	if !cfg.Application.Hello.Api.Enabled {
		return
	}
	pattern := "/hello"
	handler := handleHello(helloService, helloRepository)
	server.Handle(pattern, withTrace(pattern, handler))
}

func handleHello(helloService HelloService, helloRepository Repository) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		switch strings.ToUpper(r.Method) {
		case "GET":
			listUsers(helloRepository)(w, r)
		case "POST":
			postUsers(helloService)(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}

func withTrace(pattern string, h func(w http.ResponseWriter, r *http.Request)) http.Handler {
	return otelhttp.WithRouteTag(pattern, http.HandlerFunc(h))
}

func postUsers(helloService HelloService) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		result := helloService.Hello(r.Context(), uuid.NewString(), rand.Intn(150))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(result))
	}
}

func listUsers(helloRepository Repository) func(w http.ResponseWriter, r *http.Request) {
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
