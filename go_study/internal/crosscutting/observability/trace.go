package observability

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func WithTracing(ctx context.Context, tracerName string, spanName string, f func(context.Context)) {
	ctx, span := startSpan(ctx, tracerName, spanName)
	defer span.End()
	f(ctx)
}

func TraceWithAttr(ctx context.Context, tracerName string, spanName string, f func(context.Context, SpanModifier)) {
	ctx, span := startSpan(ctx, tracerName, spanName)
	defer span.End()
	f(ctx, OtelSpanModifier{span})
}

func WithTracingResult[T any](ctx context.Context, tracerName string, spanName string, f func(context.Context) T) T {
	ctx, span := startSpan(ctx, tracerName, spanName)
	defer span.End()
	return f(ctx)
}

func TraceWithResultAndAttr[T any](ctx context.Context, tracerName string, spanName string, f func(context.Context, SpanModifier) T) T {
	ctx, span := startSpan(ctx, tracerName, spanName)
	defer span.End()
	return f(ctx, OtelSpanModifier{span})
}

func WithTracingBiResult[T any, R any](ctx context.Context, tracerName string, spanName string, f func(context.Context) (T, R)) (T, R) {
	ctx, span := startSpan(ctx, tracerName, spanName)
	defer span.End()
	return f(ctx)
}

func startSpan(ctx context.Context, tracerName string, spanName string) (context.Context, trace.Span) {
	return otel.Tracer(tracerName).Start(ctx, fmt.Sprintf("%s.%s", tracerName, spanName))
}

type SpanModifier interface {
	HandleError(err error)
	WithAttribute(key string, value string)
}

type OtelSpanModifier struct {
	span trace.Span
}

type NoOpSpanModifier struct{}

func (sh OtelSpanModifier) HandleError(err error) {
	sh.span.RecordError(err)
}

func (sh OtelSpanModifier) WithAttribute(key string, value string) {
	sh.span.SetAttributes(attribute.String(key, value))
}

func (sh NoOpSpanModifier) HandleError(err error)                  {}
func (sh NoOpSpanModifier) WithAttribute(key string, value string) {}
