package main

import (
	"context"
	"fmt"
	"main/internal/config"
	"main/internal/hello"
	"main/internal/infra/utils/shutdown"
	"net/http"
)

func main() {
	ctx := context.Background()
	initialize(ctx)
}

func initialize(ctx context.Context) {
	cfg := config.LoadConfig()
	container := NewContainer(ctx, cfg)
	startKafkaConsumers(container)
	startProducer(container)
	startApi(container)
}

func startApi(container *Container) {
	router := http.NewServeMux()
	hello.SetupApi(router, container.Services.HelloService, container.Repositories.HelloRepository)
	srv := &http.Server{Addr: ":8080", Handler: router}

	shutdown.CreateListener(func() {
		fmt.Println("Stopping API")
		if err := srv.Shutdown(context.Background()); err != nil {
			panic(err)
		}
	})

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				fmt.Printf("Got error: %v", err)
			}
		}
	}()

	shutdown.AwaitAll()
}

func startKafkaConsumers(container *Container) {
	for _, cons := range container.Kafka.Consumers {
		if err := cons.Start(); err != nil {
			panic(err)
		}
	}
}

func startProducer(container *Container) {
	container.Workers.HelloProducerWorker.Start()
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
