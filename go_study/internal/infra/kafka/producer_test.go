package kafka

import (
	"context"
	"testing"
	"time"

	libkafka "github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

func TestToKafkaHeaders_WhenHeadersAreEmpty_ReturnsNil(t *testing.T) {
	headers := toKafkaHeaders(nil)
	if headers != nil {
		t.Fatalf("expected nil headers, got %v", headers)
	}
}

func createTopicAndWaitUntilReady(t *testing.T, cluster *libkafka.MockCluster, producer *libkafka.Producer, topic string) {
	t.Helper()

	if err := cluster.CreateTopic(topic, 1, 1); err != nil {
		t.Fatalf("failed to create topic %q: %v", topic, err)
	}

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		metadata, err := producer.GetMetadata(&topic, false, 200)
		if err == nil {
			topicMetadata, exists := metadata.Topics[topic]
			if exists && topicMetadata.Error.Code() == libkafka.ErrNoError && len(topicMetadata.Partitions) > 0 {
				return
			}
		}

		time.Sleep(20 * time.Millisecond)
	}

	t.Fatalf("topic %q did not become ready before timeout", topic)
}

func TestProducer_Produce_WhenKeyAndHeadersAreProvided_MapsMetadataIntoKafkaMessage(t *testing.T) {
	cluster, err := libkafka.NewMockCluster(1)
	if err != nil {
		t.Fatalf("failed to create mock kafka cluster: %v", err)
	}
	defer cluster.Close()

	libProducer, err := libkafka.NewProducer(&libkafka.ConfigMap{
		"bootstrap.servers":     cluster.BootstrapServers(),
		"broker.address.family": "v4",
		"message.timeout.ms":    5000,
	})
	if err != nil {
		t.Fatalf("failed to create kafka producer: %v", err)
	}
	defer libProducer.Close()

	subject := Producer{producer: libProducer}

	expectedTopic := "hello-topic"
	createTopicAndWaitUntilReady(t, cluster, libProducer, expectedTopic)
	expectedPayload := "payload"
	expectedKey := "hello-key"
	expectedHeaders := map[string]string{
		"x-tenant-id": "tenant-1",
		"x-request":   "req-123",
	}

	consumer, err := libkafka.NewConsumer(&libkafka.ConfigMap{
		"bootstrap.servers":     cluster.BootstrapServers(),
		"broker.address.family": "v4",
		"group.id":              "producer-test-group",
		"auto.offset.reset":     "earliest",
	})
	if err != nil {
		t.Fatalf("failed to create kafka consumer: %v", err)
	}
	defer consumer.Close()

	if err := consumer.SubscribeTopics([]string{expectedTopic}, nil); err != nil {
		t.Fatalf("failed to subscribe consumer: %v", err)
	}

	err = subject.Produce(context.Background(), expectedPayload, expectedTopic, expectedKey, expectedHeaders)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	consumed, err := consumer.ReadMessage(5 * time.Second)
	if err != nil {
		t.Fatalf("failed to consume message: %v", err)
	}

	if got := string(consumed.Key); got != expectedKey {
		t.Fatalf("expected key %q, got %q", expectedKey, got)
	}
	if got := string(consumed.Value); got != expectedPayload {
		t.Fatalf("expected payload %q, got %q", expectedPayload, got)
	}

	actualHeaders := make(map[string]string, len(consumed.Headers))
	for _, header := range consumed.Headers {
		actualHeaders[header.Key] = string(header.Value)
	}
	for key, value := range expectedHeaders {
		if actualHeaders[key] != value {
			t.Fatalf("expected header %q=%q, got %q", key, value, actualHeaders[key])
		}
	}
}

func TestProducer_NewProducer_WhenProducingMessage_PublishesToKafka(t *testing.T) {
	cluster, err := libkafka.NewMockCluster(1)
	if err != nil {
		t.Fatalf("failed to create mock kafka cluster: %v", err)
	}
	defer cluster.Close()

	subject, err := NewProducer(cluster.BootstrapServers())
	if err != nil {
		t.Fatalf("failed to create producer: %v", err)
	}
	defer subject.Close()

	topic := "new-producer-topic"
	createTopicAndWaitUntilReady(t, cluster, subject.producer, topic)

	consumer, err := libkafka.NewConsumer(&libkafka.ConfigMap{
		"bootstrap.servers":     cluster.BootstrapServers(),
		"broker.address.family": "v4",
		"group.id":              "producer-new-test-group",
		"auto.offset.reset":     "earliest",
	})
	if err != nil {
		t.Fatalf("failed to create kafka consumer: %v", err)
	}
	defer consumer.Close()

	if err := consumer.SubscribeTopics([]string{topic}, nil); err != nil {
		t.Fatalf("failed to subscribe consumer: %v", err)
	}

	if err := subject.Produce(context.Background(), "payload", topic, "", nil); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	consumed, err := consumer.ReadMessage(5 * time.Second)
	if err != nil {
		t.Fatalf("failed to consume message: %v", err)
	}

	if got := string(consumed.Value); got != "payload" {
		t.Fatalf("expected payload %q, got %q", "payload", got)
	}
}

func TestProducer_NewProducer_WhenClosed_StopsEventLoopAndRejectsFurtherPublishes(t *testing.T) {
	cluster, err := libkafka.NewMockCluster(1)
	if err != nil {
		t.Fatalf("failed to create mock kafka cluster: %v", err)
	}
	defer cluster.Close()

	subject, err := NewProducer(cluster.BootstrapServers())
	if err != nil {
		t.Fatalf("failed to create producer: %v", err)
	}
	subject.Close()

	topic := "closed-topic"
	deadline := time.After(5 * time.Second)
	for {
		err = subject.Produce(context.Background(), "payload", topic, "", nil)
		if err != nil {
			if kafkaErr, ok := err.(libkafka.Error); ok && kafkaErr.Code() == libkafka.ErrState {
				break
			}
			t.Fatalf("expected closed producer error, got %v", err)
		}

		select {
		case <-deadline:
			t.Fatal("expected producer to reject publishes after close")
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
}
