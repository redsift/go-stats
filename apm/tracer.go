package apm

import (
	"net"
	"net/http"

	"github.com/DataDog/dd-trace-go/tracer"
	"github.com/pierrec/xxHash/xxHash64"
)

type Tracer struct {
	*tracer.Tracer
}

// NewRootSpan creates a span with no parent. Its span id will be randomly
// assigned and its trace id calculated from request id
func (t *Tracer) NewRootSpan(name, service, resource string, reqID string) *tracer.Span {
	span := t.Tracer.NewRootSpan(name, service, resource)
	span.TraceID = xxHash64.Checksum([]byte(reqID), 0xCAFEBABE)
	return span
}

// NewTracer create a new tracer.Tracer with transport configured to send traces to addr.
// NewTracer assumes addr is "hostname:port" string, otherwise discard-all transport will be used.
func NewTracer(addr string, tags map[string]string) *Tracer {
	h, p, err := net.SplitHostPort(addr)
	var transport tracer.Transport
	if err != nil {
		transport = nullTransport{}
	} else {
		transport = tracer.NewTransport(h, p)
	}
	t := &Tracer{tracer.NewTracerTransport(transport)}
	for k, v := range tags {
		t.SetMeta(k, v)
	}
	return t
}

type nullTransport struct{}

func (_ nullTransport) SendTraces(spans [][]*tracer.Span) (*http.Response, error) {
	return nil, nil
}

func (_ nullTransport) SendServices(services map[string]tracer.Service) (*http.Response, error) {
	return nil, nil
}

func (_ nullTransport) SetHeader(key, value string) {}
