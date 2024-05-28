package victoriametrics

import (
	"time"

	"github.com/benbjohnson/clock"
)

type CollectorOption func(*Collector)

func WithClock(clock clock.Clock) CollectorOption {
	return func(c *Collector) {
		c.clock = clock
	}
}

func WithFlushInterval(interval time.Duration) CollectorOption {
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
