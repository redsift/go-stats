package victoriametrics

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"golang.org/x/exp/maps"
)

func (c *Collector) Flush() error {
	c.flush()
	return errors.Join(c.Errors()...)
}

func (c *Collector) flush() {
	c.swapLock.Lock()
	defer c.swapLock.Unlock()
	defer c.buf.Reset()

	c.lock.Lock()
	c.a, c.b = c.b, c.a
	c.lock.Unlock()

	c.render(c.b)
	maps.Clear(c.b)

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, c.buf)
	if err != nil {
		c.err(err)
		return
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		c.err(err)
		return
	}

	if res.StatusCode >= 400 {
		c.err(fmt.Errorf("API returned error status code %d: %s", res.StatusCode, res.Status))
	}
}

func (c *Collector) render(metrics map[string]*Metric) {
	for _, metric := range metrics {
		line, err := metric.MarshalJSON()
		if err != nil {
			c.err(fmt.Errorf("cannot marshal metric %q: %w", metric.Name, err))
			continue
		}
		c.buf.Write(line)
		c.buf.WriteRune('\n')
	}
}

func (c *Collector) value(stat string, value float64, tags []string) {
	if c.stopOnErr && c.errc() > 0 {
		return
	}

	slices.Sort(tags)
	key := stat + "|" + strings.Join(tags, "|")

	c.lock.Lock()
	defer c.lock.Unlock()
	m, ok := c.a[key]
	if !ok {
		if c.maxMetrics > 0 && len(c.a) >= c.maxMetrics {
			c.err(fmt.Errorf("too many metrics, discarding %q (%v)", stat, tags))
			return
		}

		m = &Metric{Name: stat}
		c.a[key] = m
		m.addLegacyTags(tags)
	}

	if c.maxPoints > 0 && len(m.Data.Values) >= c.maxPoints {
		c.err(fmt.Errorf("too many points in %q (%v)", stat, tags))
		return
	}
	m.AddValue(c.clock.Now(), value)
}

func (c *Collector) err(err error) {
	if err == nil {
		return
	}
	c.errCond.L.Lock()
	c.errors = append(c.errors, err)
	c.errCond.L.Unlock()
	c.errCond.Broadcast()
}

func (c *Collector) errc() int {
	c.errCond.L.Lock()
	defer c.errCond.L.Unlock()
	return len(c.errors)
}
