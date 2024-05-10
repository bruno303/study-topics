package worker

import (
	"encoding/json"
	"fmt"
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
	topic    string
}

func NewHelloProducerWorker(producer Producer, topic string) HelloProducerWorker {
	return HelloProducerWorker{producer: producer, topic: topic}
}

type helloKafkaMsg struct {
	Id  string `json:"id"`
	Age int    `json:"age"`
}

func (w HelloProducerWorker) Start() {
	run := true
	nextTick := time.NewTicker(5 * time.Millisecond)

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
	return w.producer.Produce(string(bytes), w.topic)
}
