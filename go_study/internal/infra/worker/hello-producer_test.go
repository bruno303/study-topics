package worker

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/bruno303/study-topics/go-study/internal/config"
	"go.uber.org/mock/gomock"
)

func TestProduce(t *testing.T) {
	producer := NewMockProducer(gomock.NewController(t))

	producer.
		EXPECT().
		Produce(gomock.Any(), gomock.Any(), gomock.Eq("topic")).
		DoAndReturn(func(ctx context.Context, msg string, topic string) error {
			return nil
		}).Times(1)

	subject := NewHelloProducerWorker(producer, config.HelloProducerConfig{
		IntervalMillis: time.Hour.Milliseconds(),
		Topic:          "topic",
		Enabled:        true,
	})

	if err := subject.produceMessage(context.Background()); err != nil {
		t.Errorf("Error '%v' was not expected", err)
	}
}

func TestProduceWithError(t *testing.T) {
	producer := NewMockProducer(gomock.NewController(t))

	producer.
		EXPECT().
		Produce(gomock.Any(), gomock.Any(), gomock.Eq("topic")).
		DoAndReturn(func(ctx context.Context, msg string, topic string) error {
			return errors.New("error")
		}).Times(1)

	subject := NewHelloProducerWorker(producer, config.HelloProducerConfig{
		IntervalMillis: time.Hour.Milliseconds(),
		Topic:          "topic",
		Enabled:        true,
	})

	err := subject.produceMessage(context.Background())

	if err == nil || err.Error() != "error" {
		t.Errorf("Error 'error' was expected")
	}
}
