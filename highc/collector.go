package highc

import (
	"time"

	"github.com/redsift/go-stats/stats"
	"github.com/redsift/go-stats/tags"
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

func (c *Collector) CountH(stat string, value float64, ts ...tags.Tag) {
	c.low.Count(stat, value, tags.List(ts).Low()...)
	c.high.Count(stat, value, tags.List(ts).All()...)
}

func (c *Collector) GaugeH(stat string, value float64, ts ...tags.Tag) {
	c.low.Gauge(stat, value, tags.List(ts).Low()...)
	c.high.Gauge(stat, value, tags.List(ts).All()...)
}

func (c *Collector) TimingH(stat string, value time.Duration, ts ...tags.Tag) {
	c.low.Timing(stat, value, tags.List(ts).Low()...)
	c.high.Timing(stat, value, tags.List(ts).All()...)
}

func (c *Collector) HistogramH(stat string, value float64, ts ...tags.Tag) {
	c.low.Histogram(stat, value, tags.List(ts).Low()...)
	c.high.Histogram(stat, value, tags.List(ts).All()...)
}

func (s *Collector) WithH(ts ...tags.Tag) stats.HighCardinalityCollector {
	return &Collector{
		low:  s.low.With(tags.List(ts).Low()...),
		high: s.high.With(tags.List(ts).All()...),
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

func (s *Collector) Unwrap() stats.HighCardinalityCollector {
	if nextH, nextL := stats.Unwrap(s.high), stats.Unwrap(s.low); nextH != s.high && nextL != s.low {
		return &Collector{
			high: stats.Unwrap(s.high),
			low:  stats.Unwrap(s.low),
		}
	}
	return s
}
