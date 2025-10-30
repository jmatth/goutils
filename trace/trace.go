// Package trace is a shim to allow for easy migration from opencensus to
// opentelemetry.
package trace

import (
	"context"

	octrace "go.opencensus.io/trace"
	otelresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

var traceProvider *sdktrace.TracerProvider
var tracer trace.Tracer

type SpanData = octrace.SpanData
type ReadOnlySpan = sdktrace.ReadOnlySpan

// Span is a wrapper type that contains either a [trace.trace] or a
// [octrace.Span].
type Span struct {
	otelSpan trace.Span
	ocSpan   *octrace.Span
}

func GetProvider() *sdktrace.TracerProvider {
	return traceProvider
}

func SetTracerWithExporter(exporter sdktrace.SpanExporter, resource *otelresource.Resource) {
	// r, err := otelresource.Merge(
	// 	otelresource.Default(),
	// 	otelresource.NewWithAttributes(semconv.SchemaURL, semconv.ServiceName("rdk")),
	// )
	// if err != nil {
	// 	panic(err)
	// }
	traceProvider = sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource),
	)
	tracer = traceProvider.Tracer("go.viam.com/rdk")
}

func Shutdown(ctx context.Context) error {
	if traceProvider != nil {
		traceProvider.ForceFlush(ctx)
		return traceProvider.Shutdown(ctx)
	}
	return nil
}

// StartSpan is a wrapper aronud [trace.Tracer.Start].
func StartSpan(ctx context.Context, name string, o ...trace.SpanStartOption) (context.Context, *Span) {
	if tracer != nil {
		ctx, span := tracer.Start(ctx, name)
		return ctx, &Span{otelSpan: span}
	}
	ctx, ocSpan := octrace.StartSpan(ctx, name)
	return ctx, &Span{ocSpan: ocSpan}
}

// FromContext is a wrapper around [trace.FromContext].
func FromContext(ctx context.Context) *Span {
	if tracer != nil {
		return &Span{otelSpan: trace.SpanFromContext(ctx)}
	}
	return &Span{ocSpan: octrace.FromContext(ctx)}
}

// NewContext is a wrapper around [trace.ContextWithSpan].
func NewContext(ctx context.Context, span *Span) context.Context {
	if span.otelSpan != nil {
		return trace.ContextWithSpan(ctx, span.otelSpan)
	}
	return octrace.NewContext(ctx, span.ocSpan)
}

func (s *Span) End() {
	if s.otelSpan != nil {
		s.otelSpan.End()
		return
	}
	s.ocSpan.End()
}

func (s *Span) AddEvent(name string, opts ...trace.EventOption) {
	if s.otelSpan != nil {
		s.otelSpan.AddEvent(name, opts...)
		return
	}
	// TODO: support annotating opencensus spans?
}