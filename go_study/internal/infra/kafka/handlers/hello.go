package handlers

import (
	"context"
	"encoding/json"

	"github.com/bruno303/study-topics/go-study/internal/application/hello"
	"github.com/bruno303/study-topics/go-study/internal/crosscutting/observability/log"
	"github.com/bruno303/study-topics/go-study/internal/crosscutting/observability/trace"
)

type HelloMessageHandler struct {
	helloService hello.HelloService
}

var _ MessageHandler = (*HelloMessageHandler)(nil)

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

	helloMsg := new(helloKafkaMsg)
	if err := json.Unmarshal([]byte(msg), helloMsg); err != nil {
		log.Log().Error(ctx, "Error during message serialization", err)
		trace.InjectError(ctx, err)
		return err
	}

	input := hello.HelloInput{
		Id:  helloMsg.Id,
		Age: helloMsg.Age,
	}
	result, err := mh.helloService.Hello(ctx, input)
	if err != nil {
		log.Log().Error(ctx, "Error during message processing", err)
		trace.InjectError(ctx, err)
		return err
	}
	log.Log().Info(ctx, "Result: %+v", result)
	return nil
}
