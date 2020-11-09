package tracer

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/api/trace"
	"go.opentelemetry.io/otel/exporters/otlp"
	"go.opentelemetry.io/otel/propagators"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv"
	"io"
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

func NewRootSpanWithID(ctx context.Context, id string, traceFlags byte) (context.Context, error) {
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
	//hash := xxHash64.Checksum([]byte(s), 0xCAFEBABE)
	//hx := hex.EncodeToString(i64tob(hash))

	////fmt.Println("hash: ", len(hash))
	//fmt.Println("i64tob(hash): ", len(i64tob(hash)))
	//fmt.Println("i32tob(hash): ", len(i32tob( xxHash32.Checksum([]byte(s), 0xCAFEBABE))))
	//fmt.Println("hx: ", len(hx))
	//fmt.Println(hx)


	//var id trace.ID
	//id, err := trace.IDFromHex(hx)
	//if err != nil {
	//	return id, err
	//}

	h := md5.New()
	_, _ = io.WriteString(h, s)
	b := h.Sum(nil)

	id := trace.ID{}
	copy(id[:],b)

	return id, nil
}

func i64tob(val uint64) []byte {
	r := make([]byte, 8)
	for i := uint64(0); i < 8; i++ {
		r[i] = byte((val >> (i * 8)) & 0xff)
	}
	return r
}

func i32tob(val uint32) []byte {
	r := make([]byte, 4)
	for i := uint32(0); i < 4; i++ {
		r[i] = byte((val >> (8 * i)) & 0xff)
	}
	return r
}