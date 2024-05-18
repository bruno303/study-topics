package worker

import (
	"encoding/json"
	"fmt"
	"main/internal/config"
	"main/internal/infra/utils/shutdown"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

type Producer interface {
	Produce(msg string, topic string) error
	Close()
}

type HelloProducerWorker struct {
	producer Producer
	cfg      config.HelloProducerConfig
}

func NewHelloProducerWorker(producer Producer, cfg config.HelloProducerConfig) HelloProducerWorker {
	return HelloProducerWorker{
		producer: producer,
		cfg:      cfg,
	}
}

type helloKafkaMsg struct {
	Id  string `json:"id"`
	Age int    `json:"age"`
}

func (w HelloProducerWorker) Start() {
	run := true
	nextTick := time.NewTicker(time.Duration(w.cfg.IntervalMillis) * time.Millisecond)

	shutdown.CreateListener(func() {
		fmt.Println("Stopping producer")
		nextTick.Stop()
		run = false
	})

	go func() {
		for range nextTick.C {
			if !run {
				return
			}
			_ = w.produceMessage()
		}
	}()

	fmt.Println("HelloProducerWorker started")
}

func (w HelloProducerWorker) produceMessage() error {
	msg := helloKafkaMsg{
		Id:  uuid.NewString(),
		Age: rand.Intn(150),
	}
	bytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return w.producer.Produce(string(bytes), w.cfg.Topic)
}
