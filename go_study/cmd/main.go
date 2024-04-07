package main

import (
	"context"
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
	hello.SetupApi(server, container.Hello)

	http.ListenAndServe(":8080", server)
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
