package victoriametrics

import (
	"strings"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/redsift/go-stats/stats"
	"github.com/stretchr/testify/require"
)

var _ stats.Collector = &Collector{}

func assertRendered(t *testing.T, collector *Collector, metrics ...string) {
	// make sure the buffer is empty
	collector.buf.Reset()

	// render to the buffer
	collector.render(collector.a)

	// ensure no error occured during rendering
	require.Len(t, collector.errors, 0)

	// ensure the metrics are all as expected, ignoring order
	result := strings.Split(strings.TrimSpace(collector.buf.String()), "\n")
	require.ElementsMatch(t, metrics, result)
}

func TestCollector(t *testing.T) {
	c := clock.NewMock()
	c.Set(time.Unix(0, 0))
	collector := NewCollector("", WithClock(c))

	// add single metric with single value
	collector.Count("test", 123, "sometag:somevalue")
	require.Len(t, collector.errors, 0)

	// ensure expected result
	assertRendered(t, collector, `{"metric":{"__name__":"test","sometag":"somevalue"},"values":[123],"timestamps":[0]}`)

	// add second value to same metric
	c.Add(time.Second)
	collector.Count("test", 42, "sometag:somevalue")
	require.Len(t, collector.errors, 0)
	require.Len(t, collector.a, 1)

	// ensure the values are added to the same metric as expected
	assertRendered(t, collector, `{"metric":{"__name__":"test","sometag":"somevalue"},"values":[123,42],"timestamps":[0,1000]}`)

	// add a second metric
	collector.Gauge("test2", 321, "sometag:somevalue")
	assertRendered(t, collector,
		`{"metric":{"__name__":"test","sometag":"somevalue"},"values":[123,42],"timestamps":[0,1000]}`,
		`{"metric":{"__name__":"test2","sometag":"somevalue"},"values":[321],"timestamps":[1000]}`,
	)

	// ensure differing tags result in new metric
	c.Add(time.Second)
	collector.Gauge("test2", 42, "sometag:someothervalue")
	assertRendered(t, collector,
		`{"metric":{"__name__":"test","sometag":"somevalue"},"values":[123,42],"timestamps":[0,1000]}`,
		`{"metric":{"__name__":"test2","sometag":"somevalue"},"values":[321],"timestamps":[1000]}`,
		`{"metric":{"__name__":"test2","sometag":"someothervalue"},"values":[42],"timestamps":[2000]}`,
	)
}
