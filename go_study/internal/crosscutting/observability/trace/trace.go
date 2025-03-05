package trace

import (
	"context"
	"errors"
	"sync"

	"github.com/bruno303/study-topics/go-study/internal/crosscutting/observability/trace/attr"
)

type (
	EndFunc     func()
	TraceKind   int
	TraceConfig struct {
		Kind      TraceKind
		TraceName string
		SpanName  string
	}
	TraceIds struct {
		TraceId string
		SpanId  string
		IsValid bool
	}
	Tracer interface {
		Trace(ctx context.Context, cfg *TraceConfig) (context.Context, EndFunc)
		ExtractTraceIds(ctx context.Context) TraceIds
		InjectAttributes(ctx context.Context, attrs ...attr.Attribute)
		InjectError(ctx context.Context, err error)
		EndTrace(ctx context.Context)
	}
)

var (
	tracer Tracer = NewNoOpTracer()
	once   sync.Once
)

const (
	_ TraceKind = iota
	TraceKindServer
	TraceKindConsumer
	TraceKindProducer
)

func GetTracer() Tracer {
	return tracer
}

func SetTracer(t Tracer) {
	once.Do(func() {
		tracer = t
	})
}

func Trace(ctx context.Context, cfg *TraceConfig) (context.Context, EndFunc) {
	return tracer.Trace(ctx, cfg)
}

func ExtractTraceIds(ctx context.Context) TraceIds {
	return tracer.ExtractTraceIds(ctx)
}

func InjectAttributes(ctx context.Context, attrs ...attr.Attribute) {
	tracer.InjectAttributes(ctx, attrs...)
}

func InjectError(ctx context.Context, err error) {
	tracer.InjectError(ctx, err)
}

func NameConfig(traceName string, spanName string) *TraceConfig {
	return &TraceConfig{TraceName: traceName, SpanName: spanName}
}

func DefaultTraceCfg() *TraceConfig {
	return &TraceConfig{
		Kind: TraceKindServer,
	}
}

func (c *TraceConfig) Validate() error {
	if c.Kind == 0 {
		c.Kind = TraceKindServer
	}
	if c.TraceName == "" {
		return errors.New("TraceName must be informed")
	}
	if c.SpanName == "" {
		return errors.New("SpanName must be informed")
	}
	return nil
}
