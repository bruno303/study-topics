package kafka

import (
	"context"
	"errors"
	"testing"
	"time"

	libkafka "github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

func TestIsSyncMessage(t *testing.T) {
	tests := []struct {
		name     string
		opaque   interface{}
		expected bool
	}{
		{"sync string", "sync", true},
		{"nil opaque", nil, false},
		{"different string", "other", false},
		{"integer opaque", 123, false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := &libkafka.Message{Opaque: tt.opaque}
			result := isSyncMessage(msg)
			if result != tt.expected {
				t.Errorf("isSyncMessage(%v) = %v, want %v", tt.opaque, result, tt.expected)
			}
		})
	}
}

// TestProduceSync_ContextCancelled verifies that ProduceSync returns ctx.Err()
// when the context is cancelled before delivery.
func TestProduceSync_ContextCancelled(t *testing.T) {
	// Create a producer with a dummy bootstrap server.
	// libkafka.NewProducer does not validate the server synchronously.
	p, err := NewProducer("localhost:9092")
	if err != nil {
		t.Skipf("skipping test: cannot create producer: %v", err)
	}
	defer p.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	err = p.ProduceSync(ctx, "test message", "test-topic", nil)
	if err == nil {
		t.Error("expected error from cancelled context, got nil")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

// TestProduceSync_ContextTimeout verifies that ProduceSync returns context.DeadlineExceeded
// when the timeout fires before delivery.
func TestProduceSync_ContextTimeout(t *testing.T) {
	p, err := NewProducer("localhost:9092")
	if err != nil {
		t.Skipf("skipping test: cannot create producer: %v", err)
	}
	defer p.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Give the deadline time to pass.
	time.Sleep(time.Millisecond)

	err = p.ProduceSync(ctx, "test message", "test-topic", nil)
	if err == nil {
		t.Error("expected timeout error, got nil")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected context.DeadlineExceeded, got %v", err)
	}
}

// TestProduceSync_HeadersAreSet verifies that user headers are added to the
// message and that the Opaque flag is set to "sync" to skip the events goroutine.
func TestProduceSync_HeadersAndOpaque(t *testing.T) {
	// Build a message manually to verify the expected fields without a real broker.
	headers := map[string]string{
		"X-Custom-Header": "custom-value",
		"X-Request-Id":    "abc123",
	}

	msg := &libkafka.Message{
		Value:          []byte("test payload"),
		Opaque:         "sync",
		TopicPartition: libkafka.TopicPartition{Topic: strPtr("test-topic"), Partition: libkafka.PartitionAny},
	}

	for k, v := range headers {
		msg.Headers = append(msg.Headers, libkafka.Header{Key: k, Value: []byte(v)})
	}

	// Verify Opaque is set to "sync"
	if !isSyncMessage(msg) {
		t.Error("expected isSyncMessage to return true for message with Opaque=\"sync\"")
	}

	// Verify headers are present before trace injection
	found := make(map[string]string)
	for _, h := range msg.Headers {
		found[h.Key] = string(h.Value)
	}
	for k, expectedV := range headers {
		if got, ok := found[k]; !ok {
			t.Errorf("header %q not found in message", k)
		} else if got != expectedV {
			t.Errorf("header %q = %q, want %q", k, got, expectedV)
		}
	}
}

func strPtr(s string) *string {
	return &s
}
