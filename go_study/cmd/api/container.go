package main

import (
	"context"

	"github.com/bruno303/study-topics/go-study/internal/config"
	"github.com/bruno303/study-topics/go-study/internal/hello"
	"github.com/bruno303/study-topics/go-study/internal/infra/database"
	"github.com/bruno303/study-topics/go-study/internal/infra/kafka"
	"github.com/bruno303/study-topics/go-study/internal/infra/kafka/handlers"
	"github.com/bruno303/study-topics/go-study/internal/infra/repository"
	"github.com/bruno303/study-topics/go-study/internal/infra/worker"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Container struct {
	Config          *config.Config
	Services        ServiceContainer
	Repositories    RepositoryContainer
	MessageHandlers MessageHandlersContainer
	Kafka           KafkaContainer
	Workers         WorkerContainer
}

type ServiceContainer struct {
	HelloService hello.HelloService
}

type RepositoryContainer struct {
	HelloRepository hello.Repository
}

type KafkaContainer struct {
	Consumers []kafka.ConsumerGroup
	Producer  kafka.Producer
}

type WorkerContainer struct {
	HelloProducerWorker worker.HelloProducerWorker
}

type MessageHandlersContainer struct {
	Hello handlers.HelloMessageHandler
}

func newServiceContainer(repositories RepositoryContainer) ServiceContainer {
	return ServiceContainer{
		HelloService: hello.NewService(repositories.HelloRepository),
	}
}

func newRepositoryContainer(ctx context.Context, pool *pgxpool.Pool) RepositoryContainer {
	return RepositoryContainer{
		HelloRepository: repository.NewHelloPgxRepository(ctx, pool),
	}
}

func newKafkaContainer(cfg *config.Config, handlers MessageHandlersContainer) KafkaContainer {
	consumers := []kafka.ConsumerGroup{}
	consumers = append(consumers, createKafkaConsumerGroup(cfg.Kafka.Consumers.GoStudy, handlers.Hello))

	kafkaProducer, err := kafka.NewProducer(cfg.Kafka.Host)
	if err != nil {
		panic(err)
	}

	return KafkaContainer{
		Consumers: consumers,
		Producer:  kafkaProducer,
	}
}

func createKafkaConsumerGroup(cfg config.KafkaConsumerConfigDetail, handler handlers.MessageHandler) kafka.ConsumerGroup {
	consumer, err := kafka.NewConsumerGroup(cfg, handler)
	if err != nil {
		panic(err)
	}
	return consumer
}

func newMessageHandlersContainer(services ServiceContainer) MessageHandlersContainer {
	return MessageHandlersContainer{
		Hello: handlers.NewHelloMessageHandler(services.HelloService),
	}
}

func newWorkerContainer(kafka KafkaContainer, cfg *config.Config) WorkerContainer {
	helloProducerWorker := worker.NewHelloProducerWorker(
		kafka.Producer,
		cfg.Workers.HelloProducer,
	)
	return WorkerContainer{
		HelloProducerWorker: helloProducerWorker,
	}
}

func NewContainer(ctx context.Context, cfg *config.Config) *Container {
	pool := database.Connect(cfg)

	repositories := newRepositoryContainer(ctx, pool)
	services := newServiceContainer(repositories)
	messageHandlers := newMessageHandlersContainer(services)
	kafka := newKafkaContainer(cfg, messageHandlers)
	worker := newWorkerContainer(kafka, cfg)

	return &Container{
		Config:          cfg,
		Services:        services,
		Repositories:    repositories,
		MessageHandlers: messageHandlers,
		Kafka:           kafka,
		Workers:         worker,
	}
}
