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

func SetupApi(server *http.ServeMux, helloService HelloService, helloRepository Repository) {
	server.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		switch strings.ToUpper(r.Method) {
		case "GET":
			listUsers(helloRepository)(w, r)
		case "POST":
			postUsers(helloService)(w, r)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
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
