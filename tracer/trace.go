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
}

// InitTracingProvider creates a new otel tracing provider and returns it with a closer.
func InitTracingProvider(collectorAddress string, serviceName string) (*Tracer, func(), error) {
	exp, err := otlp.NewExporter(
		otlp.WithInsecure(),
		otlp.WithAddress(collectorAddress),
	)
	if err != nil {
		return nil, func() {}, err
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

	t := Tracer{}
	t.Tracer = tracerProvider.Tracer("redsift/trace")

	return &t, func() {
		bsp.Shutdown() // shutdown the processor
		if err := exp.Shutdown(context.Background()); err != nil {
			fmt.Printf("Error closing tracing exporter %s: %s", err, err)
		}
	}, nil
}

// StartRootSpanWithRequestID starts a new root span where the trace and span id's are derived from the request id,
// its also adds the request id as an attribute to the span
func (t *Tracer) StartRootSpanWithRequestID(ctx context.Context, spanName string, requestID string) (context.Context, trace.Span, error) {
	traceID, spanID, err  := requestIDtoSpanIDs(requestID)
	if err != nil {
		return nil, nil, err
	}

	getIDs := func() (traceId trace.ID, spanId trace.SpanID) {
		return traceID, spanID
	}

	ctx, span := t.Start(ctx, spanName, trace.WithNewRoot(), trace.WithGetIDsFuncOption(getIDs))
	span.SetAttributes(label.String("request-id", requestID))

	return ctx, span, nil
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

func requestIDtoSpanIDs(id string) (trace.ID, trace.SpanID, error) {
	if id == "" {
		return emptyTraceID, emptySpanID, errors.New("id is empty")
	}

	traceID, err := asTraceID(id)
	if err != nil {
		return emptyTraceID, emptySpanID, err
	}

	spanID, err := asSpanID(id)
	if err != nil {
		return emptyTraceID, emptySpanID, err
	}

	return traceID, spanID, err
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
