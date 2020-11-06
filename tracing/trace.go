package tracer

import (
	"context"
	"encoding/hex"

	"github.com/pierrec/xxHash/xxHash32"
	"go.opentelemetry.io/otel/api/trace"
)

func NewRootSpanWithID(ctx context.Context, id string, traceFlags byte) (context.Context, error) {
	traceID, err := asTraceID(id)
	if err != nil {
		return nil, err
	}

	tc := trace.SpanContext{
		TraceID: traceID,
		SpanID: idGenerator.newSpanID(),
		TraceFlags: traceFlags,
	}

	return trace.ContextWithRemoteSpanContext(ctx, tc), nil
}

func asTraceID(s string) (trace.ID, error) {
	hash := xxHash32.Checksum([]byte(s), 0xCAFEBABE)
	hx := hex.EncodeToString(i32tob(hash))

	var id trace.ID
	id, err := trace.IDFromHex(hx)
	if err != nil {
		return id, err
	}

	return id, nil
}

func i32tob(val uint32) []byte {
	r := make([]byte, 4)
	for i := uint32(0); i < 4; i++ {
		r[i] = byte((val >> (8 * i)) & 0xff)
	}
	return r
}