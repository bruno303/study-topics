package tracedecorator

import (
	"context"

	"github.com/bruno303/study-topics/go-study/internal/application/hello"
	"github.com/bruno303/study-topics/go-study/internal/application/hello/models"
	"github.com/bruno303/study-topics/go-study/internal/crosscutting/observability/trace"
)

type TraceableHelloService struct {
	next hello.HelloService
}

var _ hello.HelloService = (*TraceableHelloService)(nil)

const traceName = "TraceableHelloService"

func NewTraceableHelloService(next hello.HelloService) TraceableHelloService {
	return TraceableHelloService{next: next}
}

func (t TraceableHelloService) Hello(ctx context.Context, input hello.HelloInput) (models.HelloData, error) {
	ctx, end := trace.Trace(ctx, trace.NameConfig(traceName, "Hello"))
	defer end()
	result, err := t.next.Hello(ctx, input)
	if err != nil {
		trace.InjectError(ctx, err)
		return models.HelloData{}, err
	}
	return result, nil
}

func (t TraceableHelloService) ListAll(ctx context.Context) ([]models.HelloData, error) {
	ctx, end := trace.Trace(ctx, trace.NameConfig(traceName, "ListAll"))
	defer end()
	result, err := t.next.ListAll(ctx)
	if err != nil {
		trace.InjectError(ctx, err)
		return nil, err
	}
	return result, nil
}
