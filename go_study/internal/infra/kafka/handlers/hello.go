package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"main/internal/hello"
	"main/internal/infra/observability/trace"
)

var tracer = trace.GetTracer("HelloMessageHandler")

type HelloMessageHandler struct {
	helloService hello.HelloService
}

func NewHelloMessageHandler(service hello.HelloService) HelloMessageHandler {
	return HelloMessageHandler{helloService: service}
}

type helloKafkaMsg struct {
	Id  string `json:"id"`
	Age int    `json:"age"`
}

func (mh HelloMessageHandler) Process(ctx context.Context, msg string) error {
	ctx, span := tracer.StartSpan(ctx, "Process")
	defer span.End()

	hello := new(helloKafkaMsg)
	if err := json.Unmarshal([]byte(msg), hello); err != nil {
		fmt.Printf("Error during message serialization: %v", err)
		span.SetError(err)
		return err
	}
	result := mh.helloService.Hello(ctx, hello.Id, hello.Age)
	fmt.Printf("Result: %s\n", result)
	return nil
}
