package worker

import (
	"context"
	"testing"
	"time"

	"github.com/bruno303/study-topics/go-study/internal/config"
)

type fakeProducer struct {
	lastMsg   string
	lastTopic string
}

func (p *fakeProducer) Produce(ctx context.Context, msg string, topic string) error {
	p.lastMsg = msg
	p.lastTopic = topic
	return nil
}

func (p *fakeProducer) Close() {}

func TestProduce(t *testing.T) {
	producer := &fakeProducer{}

	subject := NewHelloProducerWorker(producer, config.HelloProducerConfig{
		IntervalMillis: time.Hour.Milliseconds(),
		Topic:          "topic",
		Enabled:        true,
	})

	if err := subject.produceMessage(context.Background()); err != nil {
		t.Errorf("Error '%v' was not expected", err)
	}

	if producer.lastMsg == "" {
		t.Errorf("Msg expected")
	}
	if producer.lastTopic == "" {
		t.Errorf("Topic expected")
	}
}
