package handlers

import (
	"context"
	"encoding/json"

	"github.com/bruno303/study-topics/go-study/internal/crosscutting/observability/log"
	"github.com/bruno303/study-topics/go-study/internal/crosscutting/observability/trace"
	"github.com/bruno303/study-topics/go-study/internal/hello"
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

func (mh HelloMessageHandler) Process(ctx context.Context, msg string) error {
	ctx, end := trace.Trace(ctx, trace.NameConfig("HelloMessageHandler", "Process"))
	defer end()

	hello := new(helloKafkaMsg)
	if err := json.Unmarshal([]byte(msg), hello); err != nil {
		log.Log().Error(ctx, "Error during message serialization", err)
		trace.InjectError(ctx, err)
		return err
	}
	result := mh.helloService.Hello(ctx, hello.Id, hello.Age)
	log.Log().Info(ctx, "Result: %s", result)
	return nil
}
