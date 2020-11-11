package tracer

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/dgryski/go-metro"
	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/api/trace"
	"go.opentelemetry.io/otel/exporters/otlp"
	"go.opentelemetry.io/otel/propagators"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv"
)

// InitTracingProvider creates a new otel tracing provider and returns a closer.
func InitTracingProvider(collectorAddress string, serviceName string) (func(), error) {
	exp, err := otlp.NewExporter(
		otlp.WithInsecure(),
		otlp.WithAddress(collectorAddress),
	)
	if err != nil {
		return func() {}, err
	}

	bsp := sdktrace.NewBatchSpanProcessor(exp)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithConfig(sdktrace.Config{DefaultSampler: sdktrace.AlwaysSample()}),
		sdktrace.WithResource(resource.New(
			semconv.ServiceNameKey.String(serviceName),
		)),
		sdktrace.WithSpanProcessor(bsp),
	)

	// set global propagator to tracecontext (the default is no-op).
	global.SetTextMapPropagator(propagators.TraceContext{})
	global.SetTracerProvider(tracerProvider)

	return func() {
		bsp.Shutdown() // shutdown the processor
		if err := exp.Shutdown(context.Background()); err != nil {
			fmt.Printf("Error closing tracing exporter %s: %s", err, err)
		}
	}, nil
}

// ContextWithTraceID creates a new span context with a custom trace id.
func ContextWithTraceID(ctx context.Context, id string, traceFlags byte) (context.Context, error) {
	if id == "" {
		return nil, errors.New("trace id is empty")
	}

	traceID, err := asTraceID(id)
	if err != nil {
		return nil, err
	}

	tc := trace.SpanContext{
		TraceID:    traceID,
		SpanID:     idGenerator.newSpanID(),
		TraceFlags: traceFlags,
	}

	return trace.ContextWithRemoteSpanContext(ctx, tc), nil
}

// ContextWithSpan creates a new span context with a custom trace id.
func ContextWithIDs(ctx context.Context, id string, traceFlags byte) (context.Context, error) {
	if id == "" {
		return nil, errors.New("trace id is empty")
	}

	traceID, err := asTraceID(id)
	if err != nil {
		return nil, err
	}

	spanID, err := asSpanID(id)
	if err != nil {
		return nil, err
	}

	sc := trace.SpanContext{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: traceFlags,
	}

	return trace.ContextWithRemoteSpanContext(ctx, sc), nil
}

func asTraceID(s string) (trace.ID, error) {
	h1, h2 := metro.Hash128([]byte(s), 0xCAFEBABE)

	id := trace.ID{}
	binary.LittleEndian.PutUint64(id[:8], h1)
	binary.LittleEndian.PutUint64(id[8:], h2)

	return id, nil
}

func asSpanID(s string) (trace.SpanID, error) {
	h := metro.Hash64([]byte(s), 0xCAFEBABE)

	id := trace.SpanID{}
	binary.LittleEndian.PutUint64(id[:8], h)

	return id, nil
}
