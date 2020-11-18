package tracer

import (
	"context"
	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/api/trace"
	"go.opentelemetry.io/otel/exporters/otlp"
	"go.opentelemetry.io/otel/label"
	"go.opentelemetry.io/otel/propagators"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv"
)

var (
	emptyTraceID = trace.ID{}
	emptySpanID = trace.SpanID{}
)

type Tracer struct {
	trace.Tracer

	batchSpanProcessor *sdktrace.BatchSpanProcessor
	exporter *otlp.Exporter
}

// InitTracingProvider creates a new otel tracing provider and returns it.
func InitTracingProvider(collectorAddress string, serviceName string) (*Tracer, error) {
	exp, err := otlp.NewExporter(
		otlp.WithInsecure(),
		otlp.WithAddress(collectorAddress),
	)
	if err != nil {
		return nil, err
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

	return &Tracer{
		Tracer: tracerProvider.Tracer("redsift/trace"),
		batchSpanProcessor: bsp,
		exporter:exp,
	}, nil
}

// StartRootSpanWithRequestID starts a new root span where the trace and span id's are derived from the request id,
// its also adds the request id as an attribute to the span
func (t *Tracer) StartRootSpanWithRequestID(ctx context.Context, spanName string, requestID string) (context.Context, trace.Span, error) {
	generator, err := newIDGenerator(requestID)
	if err != nil {
		return nil, nil, err
	}

	ctx, span := t.Start(ctx, spanName, trace.WithNewRoot(), trace.WithIDGenerator(generator))
	span.SetAttributes(label.String("request-id", requestID))

	return ctx, span, nil
}

// Close closes the exporter and span processor
func (t *Tracer) Close(ctx context.Context) error {
	t.batchSpanProcessor.Shutdown()
	return t.exporter.Shutdown(ctx)
}


// ContextWithRemoteSpanIDs creates a new span context with a remote span and trace id's set to the one provided.
func ContextWithRemoteSpanIDs(ctx context.Context, id string, traceFlags byte) (context.Context, error) {
	traceID, spanID, err := requestIDtoSpanIDs(id)
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

