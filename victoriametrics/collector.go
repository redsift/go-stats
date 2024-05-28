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
		closed:     make(chan struct{}),
		a:          make(map[string]*Metric),
		b:          make(map[string]*Metric),
		buf:        &bytes.Buffer{},
	}

	for _, opt := range options {
		opt(c)
	}

	return c
}

type Collector struct {
	endpoint   string
	httpClient *http.Client
	timeout    time.Duration
	clock      clock.Clock

	closed   chan struct{}      // closed on close
	lock     sync.Mutex         // guarding the current write buffer
	a        map[string]*Metric // current write buffer, guarded by lock
	swapLock sync.Mutex         // guarding the next buffer and error list
	b        map[string]*Metric // next buffer, guarded by swapLock
	buf      *bytes.Buffer      // write buffer
	errors   []error            // push errors, guarded by swapLock
}

func (c *Collector) Inform(title, text string, tags ...string) {
	c.swapLock.Lock()
	c.errors = append(c.errors, fmt.Errorf("unhandled call to Inform(%q, %q, %v)", title, text, tags))
	c.swapLock.Unlock()
}

func (c *Collector) Error(err error, tags ...string) {
	c.swapLock.Lock()
	c.errors = append(c.errors, fmt.Errorf("unhandled call to Error(%w, %v)", err, tags))
	c.swapLock.Unlock()
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
	c.swapLock.Lock()
	errors, c.errors = c.errors, nil
	c.swapLock.Unlock()
	return
}

func (c *Collector) With(tags ...string) stats.Collector {
	return stats.NewWithCollector(c, tags...)
}
