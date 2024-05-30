package trace

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func GetTracer(tracerName string) Tracer {
	return defaultTracer{tracer: otel.Tracer(tracerName)}
}

// Tracer
type Tracer interface {
	StartSpan(ctx context.Context, spanName string) (context.Context, Span)
}

type defaultTracer struct {
	tracer trace.Tracer
}

func (t defaultTracer) StartSpan(ctx context.Context, spanName string) (context.Context, Span) {
	tCtx, span := t.tracer.Start(ctx, spanName)
	return tCtx, defaultSpan{span: span}
}

// Span
type Span interface {
	End()
	GetTraceId() string
	GetSpanId() string
	SetAttributes(attrs ...SpanAttribute)
	SetError(err error)
}

type SpanAttribute struct {
	Name  string
	Value string
}

func Attribute(name string, value string) SpanAttribute {
	return SpanAttribute{name, value}
}

type defaultSpan struct {
	span trace.Span
}

func (s defaultSpan) End() {
	s.span.End()
}

func (s defaultSpan) GetTraceId() string {
	return s.span.SpanContext().TraceID().String()
}

func (s defaultSpan) GetSpanId() string {
	return s.span.SpanContext().SpanID().String()
}

func (s defaultSpan) SetAttributes(attrs ...SpanAttribute) {
	keyValues := make([]attribute.KeyValue, 0, len(attrs))
	for _, attr := range attrs {
		attr := attribute.KeyValue{Key: attribute.Key(attr.Name), Value: attribute.StringValue(attr.Value)}
		keyValues = append(keyValues, attr)
	}
	s.span.SetAttributes(keyValues...)
}

func (s defaultSpan) SetError(err error) {
	s.span.RecordError(err)
}
