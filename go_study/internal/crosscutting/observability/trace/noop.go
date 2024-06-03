package trace

import (
	"context"
	"main/internal/crosscutting/observability/trace/attr"
)

type NoOpTracer struct{}

func NewNoOpTracer() NoOpTracer {
	return NoOpTracer{}
}

func (t NoOpTracer) Trace(ctx context.Context, cfg *TraceConfig) (context.Context, EndFunc) {
	return ctx, func() {}
}

func (t NoOpTracer) ExtractTraceIds(ctx context.Context) TraceIds {
	return TraceIds{
		TraceId: "",
		SpanId:  "",
		IsValid: false,
	}
}

func (t NoOpTracer) InjectAttributes(ctx context.Context, attrs ...attr.Attribute) {}

func (t NoOpTracer) InjectError(ctx context.Context, err error) {}
