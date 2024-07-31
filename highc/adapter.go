package highc

import (
	"time"

	"github.com/redsift/go-stats/stats"
)

func New(low, high stats.Collector) *Collector {
	return &Collector{
		low:  low,
		high: high,
	}
}

type Collector struct {
	low, high stats.Collector
}

func (c *Collector) CountH(stat string, value float64, low, high []string) {
	c.low.Count(stat, value, low...)
	c.high.Count(stat, value, append(low, high...)...)
}

func (c *Collector) GaugeH(stat string, value float64, low, high []string) {
	c.low.Gauge(stat, value, low...)
	c.high.Gauge(stat, value, append(low, high...)...)
}

func (c *Collector) TimingH(stat string, value time.Duration, low, high []string) {
	c.low.Timing(stat, value, low...)
	c.high.Timing(stat, value, append(low, high...)...)
}

func (c *Collector) HistogramH(stat string, value float64, low, high []string) {
	c.low.Histogram(stat, value, low...)
	c.high.Histogram(stat, value, append(low, high...)...)
}

func (s *Collector) WithH(low, high []string) stats.HighCardinalityCollector {
	return &Collector{
		low:  s.low.With(low...),
		high: s.high.With(low...).With(high...),
	}
}

func (s *Collector) Inform(title, text string, tags ...string) {
	s.low.Inform(title, text, tags...)
	s.high.Inform(title, text, tags...)
}

func (s *Collector) Error(err error, tags ...string) {
	s.low.Error(err, tags...)
	s.high.Error(err, tags...)
}

func (s *Collector) Count(stat string, value float64, tags ...string) {
	s.low.Count(stat, value, tags...)
	s.high.Count(stat, value, tags...)
}

func (s *Collector) Gauge(stat string, value float64, tags ...string) {
	s.low.Gauge(stat, value, tags...)
	s.high.Gauge(stat, value, tags...)
}

func (s *Collector) Timing(stat string, value time.Duration, tags ...string) {
	s.low.Timing(stat, value, tags...)
	s.high.Timing(stat, value, tags...)
}

func (s *Collector) Histogram(stat string, value float64, tags ...string) {
	s.low.Histogram(stat, value, tags...)
	s.high.Histogram(stat, value, tags...)
}

func (s *Collector) Close() {
	s.low.Close()
	s.high.Close()
}

func (s *Collector) Tags() []string {
	return nil
}

func (s *Collector) With(tags ...string) stats.Collector {
	return &Collector{
		low:  s.low.With(tags...),
		high: s.high.With(tags...),
	}
}

func (s *Collector) High() stats.Collector {
	return s.high
}

func (s *Collector) Low() stats.Collector {
	return s.low
}
