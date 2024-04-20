package main

import (
	"context"
	"main/internal/config"
	"main/internal/hello"
	"main/internal/infra/database"
	"main/internal/infra/kafka"
	"main/internal/infra/kafka/handlers"
	"main/internal/infra/repository"
	"main/internal/infra/worker"

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
		HelloRepository: repository.NewHelloRepository(ctx, pool),
	}
}

func newKafkaContainer(cfg *config.Config, handlers MessageHandlersContainer) KafkaContainer {
	consumers := []kafka.ConsumerGroup{}
	consumers = append(consumers, createKafkaConsumerGroup(&kafka.Config{
		Host:         cfg.Kafka.Host,
		Topic:        cfg.Kafka.Consumers.GoStudy.Topic,
		GroupId:      cfg.Kafka.Consumers.GoStudy.GroupId,
		QntConsumers: cfg.Kafka.Consumers.GoStudy.QntConsumers,
		Handler:      handlers.Hello,
	}))

	kafkaProducer, err := kafka.NewProducer(cfg.Kafka.Host)
	if err != nil {
		panic(err)
	}

	return KafkaContainer{
		Consumers: consumers,
		Producer:  kafkaProducer,
	}
}

func createKafkaConsumerGroup(cfg *kafka.Config) kafka.ConsumerGroup {
	consumer, err := kafka.NewConsumerGroup(cfg)
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

func newWorkerContainer(kafka KafkaContainer, cfg *config.KafkaConfig) WorkerContainer {
	helloProducerWorker := worker.NewHelloProducerWorker(kafka.Producer, cfg.Consumers.GoStudy.Topic)
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
	worker := newWorkerContainer(kafka, cfg.Kafka)

	return &Container{
		Config:          cfg,
		Services:        services,
		Repositories:    repositories,
		MessageHandlers: messageHandlers,
		Kafka:           kafka,
		Workers:         worker,
	}
}
