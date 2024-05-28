package victoriametrics

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"golang.org/x/exp/maps"
)

func (c *Collector) Flush() {
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
		c.errors = append(c.errors, err)
		return
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		c.errors = append(c.errors, err)
		return
	}

	if res.StatusCode >= 400 {
		c.errors = append(c.errors, fmt.Errorf("API returned error status code %d: %s", res.StatusCode, res.Status))
	}
}

func (c *Collector) render(metrics map[string]*Metric) {
	for _, metric := range metrics {
		line, err := metric.MarshalJSON()
		if err != nil {
			c.errors = append(c.errors, fmt.Errorf("cannot marshal metric %q: %w", metric.Name, err))
			continue
		}
		c.buf.Write(line)
		c.buf.WriteRune('\n')
	}
}

func (c *Collector) value(stat string, value float64, tags []string) {
	slices.Sort(tags)
	key := stat + "|" + strings.Join(tags, "|")

	c.lock.Lock()
	defer c.lock.Unlock()
	m, ok := c.a[key]
	if !ok {
		m = &Metric{Name: stat}
		c.a[key] = m
		m.addLegacyTags(tags)
	}
	m.AddValue(c.clock.Now(), value)
}
