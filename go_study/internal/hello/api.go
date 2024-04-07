package hello

import (
	"fmt"
	"math/rand"
	"net/http"
	"strings"

	"github.com/google/uuid"
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

func SetupApi(server *http.ServeMux, container Container) {
	server.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		switch strings.ToUpper(r.Method) {
		case "GET":
			listUsers(container)(w, r)
		case "POST":
			postUsers(container)(w, r)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
}

func postUsers(container Container) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		result := container.Service.Hello(r.Context(), uuid.NewString(), rand.Intn(150))
		w.Write([]byte(result))
		w.WriteHeader(200)
	}
}

func listUsers(container Container) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		result := container.Repository.ListAll(r.Context())
		response := ""
		for i, value := range result {
			response += value.ToString()
			if i < len(result)-1 {
				response += ", "
			}
		}
		w.Write([]byte(response))
		w.WriteHeader(200)
	}
}
