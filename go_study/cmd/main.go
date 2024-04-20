package main

import (
	"context"
	"fmt"
	"main/internal/config"
	"main/internal/hello"
	"net/http"
)

func main() {
	ctx := context.Background()
	initialize(ctx)
}

func initialize(ctx context.Context) {
	cfg := config.LoadConfig()
	container := NewContainer(ctx, cfg)
	go startKafkaConsumers(container)
	startProducer(container)

	server := http.NewServeMux()
	hello.SetupApi(server, container.Services.HelloService, container.Repositories.HelloRepository)

	fmt.Println("Application started")
	if err := http.ListenAndServe(":8080", server); err != nil {
		fmt.Printf("Got error: %v", err)
	}
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
