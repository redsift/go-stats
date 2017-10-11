package apm

import (
	"context"
	"net"
	"net/http"

	"github.com/DataDog/dd-trace-go/tracer"
	"github.com/pierrec/xxHash/xxHash64"
)

type Tracer struct {
	*tracer.Tracer
}

// AsTraceID is a pure function which converts string s to uint64.
func AsTraceID(s string) uint64 {
	return xxHash64.Checksum([]byte(s), 0xCAFEBABE)
}

// NewRootSpanWithID creates a span with no parent.
// Its span and trace ids set to given id.
func (t *Tracer) NewRootSpanWithID(op, service, resource string, id uint64) *tracer.Span {
	span := t.Tracer.NewRootSpan(op, service, resource)
	span.TraceID = id
	span.SpanID = id
	return span
}

// NewRootSpan creates a span with no parent. Its span id will be randomly
// assigned and its parent and trace ids set to given id.
func (t *Tracer) NewRootSpanWithRemoteID(op, service, resource string, id uint64) *tracer.Span {
	span := t.Tracer.NewRootSpan(op, service, resource)
	span.TraceID = id
	span.ParentID = id
	return span
}

type TracerOption func(*Tracer)

// WithMeta is an option for setting meta on tracer
func WithMeta(tags map[string]string) TracerOption {
	return func(t *Tracer) {
		for k, v := range tags {
			t.SetMeta(k, v)
		}
	}
}

// NewTracer create a new tracer.Tracer with transport configured to send traces to addr.
// NewTracer assumes addr is "hostname:port" string, otherwise discard-all transport will be used.
func NewTracer(addr string, opts ...TracerOption) *Tracer {
	h, p, err := net.SplitHostPort(addr)
	var transport tracer.Transport
	if err != nil {
		transport = nullTransport{}
	} else {
		transport = tracer.NewTransport(h, p)
	}
	t := &Tracer{tracer.NewTracerTransport(transport)}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

// NewChildSpanFromContext will create a child span of the span contained in
// the given context. If the context contains no span, a span with
// no service or resource will be returned.
func NewChildSpanFromContext(name string, ctx context.Context) *tracer.Span {
	span, _ := tracer.SpanFromContext(ctx)
	if span == nil {
		return tracer.DefaultTracer.NewChildSpan(name, span)
	}
	return span.Tracer().NewChildSpan(name, span)
}

type nullTransport struct{}

func (_ nullTransport) SendTraces(spans [][]*tracer.Span) (*http.Response, error) {
	return nil, nil
}

func (_ nullTransport) SendServices(services map[string]tracer.Service) (*http.Response, error) {
	return nil, nil
}

func (_ nullTransport) SetHeader(key, value string) {}
