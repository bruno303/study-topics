package worker

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
)

type Producer interface {
	Produce(msg string, topic string) error
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
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		nextTick := time.After(1 * time.Second)
		for {
			select {
			case <-stopChan:
				fmt.Println("Stopping producer")
				return
			case <-nextTick:
				fmt.Println("Producing message")
				_ = w.produceMessage()
				nextTick = time.After(1 * time.Second)
			}
		}
	}()
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
