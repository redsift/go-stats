package tracer

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"io"

	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/api/trace"
	"go.opentelemetry.io/otel/exporters/otlp"
	"go.opentelemetry.io/otel/propagators"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv"

)

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

func NewSpanWithTraceID(ctx context.Context, id string, traceFlags byte) (context.Context, error) {
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

func asTraceID(s string) (trace.ID, error) {
	h := md5.New()
	_, _ = io.WriteString(h, s)
	b := h.Sum(nil)

	id := trace.ID{}
	copy(id[:],b)

	return id, nil
}
