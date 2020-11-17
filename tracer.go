package go_stats

import (
	"context"

	"go.opentelemetry.io/otel/api/trace"
)

type Tracer interface {
	trace.Tracer

	StartRootSpanWithRequestID(ctx context.Context, spanName string, requestID string) (context.Context, trace.Span, error)
	Close(ctx context.Context) error
}

