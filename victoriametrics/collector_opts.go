package victoriametrics

import (
	"net/http"
	"time"

	"github.com/benbjohnson/clock"
)

type CollectorOption func(*Collector)

func WithClock(clock clock.Clock) CollectorOption {
	return func(c *Collector) {
		c.clock = clock
	}
}

func WithBackgroundFlushLoop(interval time.Duration) CollectorOption {
	return func(c *Collector) {
		ticker := c.clock.Ticker(interval)
		go func() {
			defer ticker.Stop()
			for {
				select {
				case <-c.closed:
					return
				case <-ticker.C:
					c.Flush()
				}
			}
		}()
	}
}

func WithFlushTimeout(timeout time.Duration) CollectorOption {
	return func(c *Collector) {
		c.timeout = timeout
	}
}

type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

func WithHTTPClient(client HTTPClient) CollectorOption {
	return func(c *Collector) {
		c.httpClient = client
	}
}

func WithMaxMetrics(m int) CollectorOption {
	return func(c *Collector) {
		c.maxMetrics = m
	}
}

func WithoutMetricsLimit() CollectorOption {
	return func(c *Collector) {
		c.maxMetrics = 0
	}
}

func WithPointsLimit(m int) CollectorOption {
	return func(c *Collector) {
		c.maxPoints = m
	}
}

func WithoutPointsLimit() CollectorOption {
	return func(c *Collector) {
		c.maxPoints = 0
	}
}

func WithStopOnError() CollectorOption {
	return func(c *Collector) {
		c.stopOnErr = true
	}
}

func WithoutStopOnError() CollectorOption {
	return func(c *Collector) {
		c.stopOnErr = false
	}
}

func WithErrorChannel(errors chan<- error) CollectorOption {
	return func(c *Collector) {
		defer close(errors)

		go func() {
			var err error
			for {
				c.errCond.L.Lock()
				for len(c.errors) == 0 {
					wait := make(chan struct{})

					go func() {
						c.errCond.Wait()
						close(wait)
					}()

					select {
					case <-c.closed:
						return
					case <-wait:
						continue
					}
				}

				err, c.errors = c.errors[0], c.errors[1:]
				c.errCond.L.Unlock()

				select {
				case <-c.closed:
					return
				case errors <- err:
				}
			}
		}()
	}
}
