package otel

import (
	"context"
	"fmt"
	"time"

	"github.com/bruno303/study-topics/go-study/pkg/utils/array"

	"github.com/bruno303/study-topics/go-study/internal/crosscutting/observability/trace"
	"github.com/bruno303/study-topics/go-study/internal/crosscutting/observability/trace/attr"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	tracelib "go.opentelemetry.io/otel/trace"
)

type OtelTracerAdapter struct{}

func NewOtelTracerAdapter() OtelTracerAdapter {
	return OtelTracerAdapter{}
}

func (t OtelTracerAdapter) Trace(ctx context.Context, cfg *trace.TraceConfig) (context.Context, trace.EndFunc) {
	if cfg == nil {
		cfg = trace.DefaultTraceCfg()
	}
	cfg.Validate()

	ctx, _ = startSpan(ctx, cfg.TraceName, cfg.SpanName)
	return ctx, func() { EndTrace(ctx) }
}

func (t OtelTracerAdapter) ExtractTraceIds(ctx context.Context) trace.TraceIds {
	span := tracelib.SpanFromContext(ctx)
	return trace.TraceIds{
		TraceId: span.SpanContext().TraceID().String(),
		SpanId:  span.SpanContext().SpanID().String(),
		IsValid: span.SpanContext().IsValid(),
	}
}

func (t OtelTracerAdapter) InjectAttributes(ctx context.Context, attrs ...attr.Attribute) {
	span := tracelib.SpanFromContext(ctx)
	if span == nil {
		return
	}
	otelAttrs := array.Map(attrs, func(a attr.Attribute) attribute.KeyValue {
		return attribute.String(a.Key, a.Value)
	})
	span.SetAttributes(otelAttrs...)
}

func (t OtelTracerAdapter) InjectError(ctx context.Context, err error) {
	span := tracelib.SpanFromContext(ctx)
	if span == nil {
		return
	}
	span.RecordError(err)
}

func (t OtelTracerAdapter) EndTrace(ctx context.Context) {
	EndTrace(ctx)
}

func startSpan(ctx context.Context, tracerName string, spanName string) (context.Context, tracelib.Span) {
	return otel.Tracer(tracerName).Start(
		ctx,
		fmt.Sprintf("%s.%s", tracerName, spanName),
		tracelib.WithTimestamp(time.Now()),
	)
}

func EndTrace(ctx context.Context) {
	span := tracelib.SpanFromContext(ctx)
	if span == nil {
		return
	}
	span.End()
}
