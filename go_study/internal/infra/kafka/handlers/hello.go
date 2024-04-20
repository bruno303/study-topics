package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"main/internal/hello"
)

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

func (mh HelloMessageHandler) Process(msg string) error {
	hello := new(helloKafkaMsg)
	if err := json.Unmarshal([]byte(msg), hello); err != nil {
		fmt.Printf("Error during message serialization: %v", err)
		return err
	}
	ctx := context.Background()
	result := mh.helloService.Hello(ctx, hello.Id, hello.Age)
	fmt.Printf("Result: %s\n", result)
	return nil
}
