package tracer

import (
	crand "crypto/rand"
	"encoding/binary"
	"math/rand"
	"sync"

	"go.opentelemetry.io/otel/api/trace"
)

// taken from go.opentelemetry.io/otel
type defaultIDGenerator struct {
	sync.Mutex
	randSource *rand.Rand
}

var idGenerator = defIDGenerator()

func defIDGenerator() *defaultIDGenerator {
	gen := &defaultIDGenerator{}
	var rngSeed int64
	_ = binary.Read(crand.Reader, binary.LittleEndian, &rngSeed)
	gen.randSource = rand.New(rand.NewSource(rngSeed))
	return gen
}

// newSpanID returns a non-zero span ID from a randomly-chosen sequence.
func (gen *defaultIDGenerator) newSpanID() trace.SpanID {
	gen.Lock()
	defer gen.Unlock()
	sid := trace.SpanID{}
	gen.randSource.Read(sid[:])
	return sid
}
