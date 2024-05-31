package main

import (
	"context"
	"fmt"
	"main/internal/config"
	"main/internal/hello"
	"main/internal/infra/utils/shutdown"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func main() {
	ctx := context.Background()
	initialize(ctx)
}

func initialize(ctx context.Context) {
	cfg := config.LoadConfig()
	otelShutdown := configureObservability(ctx, cfg)
	defer func() {
		otelShutdown(context.Background())
	}()

	container := NewContainer(ctx, cfg)

	startKafkaConsumers(container)
	startProducer(container)
	startApi(container)
}

func configureObservability(ctx context.Context, cfg *config.Config) func(context.Context) error {
	otelShutdown, err := SetupOTelSDK(ctx, cfg)
	if err != nil {
		panic(err)
	}
	return otelShutdown
}

func startApi(container *Container) {
	router := http.NewServeMux()
	hello.SetupApi(router, container.Services.HelloService, container.Repositories.HelloRepository)
	srv := &http.Server{Addr: ":8080", Handler: otelhttp.NewHandler(router, "/")}

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
