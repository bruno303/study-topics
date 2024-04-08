package main

import (
	"context"
	"fmt"
	"main/internal/hello"
	"net/http"
)

func main() {
	ctx := context.Background()
	initialize(ctx)
}

func initialize(ctx context.Context) {
	// channel = make(chan struct{})
	container := NewContainer(ctx)
	server := http.NewServeMux()
	hello.SetupApi(server, container.Services.HelloService, container.Repositories.HelloRepository)

	fmt.Println("Application started")
	if err := http.ListenAndServe(":8080", server); err != nil {
		fmt.Printf("Got error: %v", err)
	}
}

// func main() {
// 	initialize()
// 	wg := sync.WaitGroup{}
// 	wg.Add(100)
// 	defer wg.Wait()

// 	for i := 0; i < 100; i++ {
// 		age := i
// 		go func() {
// 			result := Container.Hello.Service.Hello(context.Background(), uuid.NewString(), age)
// 			fmt.Println(result)
// 			wg.Done()
// 		}()
// 	}
// }
