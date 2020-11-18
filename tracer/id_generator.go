package tracer

import (
	"encoding/binary"
	"errors"

	"github.com/dgryski/go-metro"

	"go.opentelemetry.io/otel/api/trace"
)

// IDGenerator is a new trace IDGenerator that provides predefined ids instead randomly generated ones
type IDGenerator struct {
	traceID trace.ID
	spanID trace.SpanID
}

func newIDGenerator(id string) (trace.IDGenerator, error) {
	g := &IDGenerator{}

	var err error
	g.traceID, g.spanID, err = requestIDtoSpanIDs(id)
	if err != nil {
		return nil, err
	}

	return g, nil
}

func (g *IDGenerator) NewTraceID() trace.ID  {
	return  g.traceID
}

func (g *IDGenerator) NewSpanID() trace.SpanID  {
	return  g.spanID
}

func requestIDtoSpanIDs(id string) (trace.ID, trace.SpanID, error) {
	if id == "" {
		return emptyTraceID, emptySpanID, errors.New("id is empty")
	}

	// todo check length

	return asTraceID(id), asSpanID(id), nil
}


func asTraceID(s string) trace.ID {
	h1, h2 := metro.Hash128([]byte(s), 0xCAFEBABE)

	id := trace.ID{}
	binary.LittleEndian.PutUint64(id[:8], h1)
	binary.LittleEndian.PutUint64(id[8:], h2)

	return id
}

func asSpanID(s string) trace.SpanID {
	h := metro.Hash64([]byte(s), 0xCAFEBABE)

	id := trace.SpanID{}
	binary.LittleEndian.PutUint64(id[:8], h)

	return id
}
