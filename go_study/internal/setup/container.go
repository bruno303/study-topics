package setup

import (
	"context"

	"github.com/bruno303/study-topics/go-study/internal/application/hello"
	"github.com/bruno303/study-topics/go-study/internal/application/transaction"
	"github.com/bruno303/study-topics/go-study/internal/config"
	"github.com/bruno303/study-topics/go-study/internal/infra/database"
	"github.com/bruno303/study-topics/go-study/internal/infra/kafka"
	"github.com/bruno303/study-topics/go-study/internal/infra/kafka/handlers"
	"github.com/bruno303/study-topics/go-study/internal/infra/observability/otel/tracedecorator"
	"github.com/bruno303/study-topics/go-study/internal/infra/repository"
	"github.com/bruno303/study-topics/go-study/internal/infra/worker"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Container struct {
	Config          *config.Config
	Services        ServiceContainer
	MessageHandlers MessageHandlersContainer
	Kafka           KafkaContainer
	Workers         WorkerContainer
	Repository      RepositoryContainer
}

type ServiceContainer struct {
	HelloService hello.HelloService
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

type RepositoryContainer struct {
	UnitOfWork transaction.UnitOfWork
}

func newServiceContainer(repoContainer RepositoryContainer) ServiceContainer {
	helloService := hello.NewService(repoContainer.UnitOfWork)

	return ServiceContainer{
		HelloService: tracedecorator.NewTraceableHelloService(helloService),
	}
}

func newRepositoryContainer(cfg *config.Config, pool *pgxpool.Pool) RepositoryContainer {
	var unitOfWork transaction.UnitOfWork

	switch cfg.Database.Driver {
	case config.DatabaseDriverMemDB:
		sharedMemDbStorage := repository.NewMemDbStorage()
		unitOfWork = repository.NewMemDbUnitOfWork(sharedMemDbStorage)
	default:
		txConfig := &repository.PgxUnitOfWorkConfig{Pool: pool}
		unitOfWork = repository.NewPgxUnitOfWork(txConfig)
	}

	return RepositoryContainer{
		UnitOfWork: unitOfWork,
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
	var pool *pgxpool.Pool
	if cfg.Database.Driver == config.DatabaseDriverPGXPool {
		pool = database.Connect(cfg)
	}

	repos := newRepositoryContainer(cfg, pool)
	services := newServiceContainer(repos)
	messageHandlers := newMessageHandlersContainer(services)
	kafka := newKafkaContainer(cfg, messageHandlers)
	worker := newWorkerContainer(kafka, cfg)

	return &Container{
		Config:          cfg,
		Services:        services,
		MessageHandlers: messageHandlers,
		Kafka:           kafka,
		Workers:         worker,
		Repository:      repos,
	}
}
