package victoriametrics

import (
	"bytes"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/redsift/go-stats/stats"
)

func NewCollector(endpoint string, options ...CollectorOption) *Collector {
	c := &Collector{
		endpoint:   endpoint,
		httpClient: http.DefaultClient,
		timeout:    5 * time.Second,
		clock:      clock.New(),
		maxMetrics: 1000,
		maxPoints:  100,
		stopOnErr:  true,
		closed:     make(chan struct{}),
		a:          make(map[string]*Metric),
		b:          make(map[string]*Metric),
		buf:        &bytes.Buffer{},
		errCond:    sync.Cond{L: &sync.Mutex{}},
	}

	for _, opt := range options {
		opt(c)
	}

	return c
}

type Collector struct {
	endpoint   string        // victoriametrics endpoint
	httpClient HTTPClient    // http client used to post the metrics
	timeout    time.Duration // http timeout
	clock      clock.Clock   // makes clock behavior deterministic in tests
	maxMetrics int           // if > 0 maximum number of metrics allowed
	maxPoints  int           // if > 0 maximum number of data points allowed in a metric
	stopOnErr  bool          // if true we don't collect any metrics after an error, until the errors are collected

	closed chan struct{} // closed on close, used to stop flush loop

	lock sync.Mutex         // guarding the current write buffer
	a    map[string]*Metric // current write buffer, guarded by lock

	swapLock sync.Mutex         // guarding the next buffer
	b        map[string]*Metric // next buffer, guarded by swapLock
	buf      *bytes.Buffer      // write buffer

	errCond sync.Cond // guarding the errors array
	errors  []error   // push errors, guarded by errCond.L
}

func (c *Collector) Inform(title, text string, tags ...string) {
	c.err(fmt.Errorf("unhandled call to Inform(%q, %q, %v)", title, text, tags))
}

func (c *Collector) Error(err error, tags ...string) {
	c.err(fmt.Errorf("unhandled call to Error(%w, %v)", err, tags))
}

func (c *Collector) Count(stat string, count float64, tags ...string) {
	c.value(stat, count, tags)
}

func (c *Collector) Gauge(stat string, value float64, tags ...string) {
	c.value(stat, value, tags)
}

func (c *Collector) Timing(stat string, value time.Duration, tags ...string) {
	c.value(stat, value.Seconds(), tags)
}

func (c *Collector) Histogram(stat string, value float64, tags ...string) {
	c.value(stat, value, tags)
}

func (c *Collector) Close() {
	close(c.closed)
	c.Flush()
}

func (c *Collector) Tags() []string {
	return nil
}

func (c *Collector) Errors() (errors []error) {
	c.errCond.L.Lock()
	errors, c.errors = c.errors, nil
	c.errCond.L.Unlock()
	return
}

func (c *Collector) With(tags ...string) stats.Collector {
	return stats.NewWithCollector(c, tags...)
}
