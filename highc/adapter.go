package highc

import (
	"time"

	"github.com/redsift/go-stats/stats"
	"github.com/redsift/go-stats/tags"
)

func NewAdapter(hc stats.HighCardinalityCollector) *Adapter {
	return &Adapter{
		HighCardinalityCollector: hc,
	}
}

type Adapter struct {
	stats.HighCardinalityCollector
}

func (s *Adapter) Inform(title, text string, ts ...tags.Tag) {
	s.Low().Inform(title, text, tags.List(ts).Low()...)
	s.High().Inform(title, text, tags.List(ts).All()...)
}

func (s *Adapter) Error(err error, ts ...tags.Tag) {
	s.Low().Error(err, tags.List(ts).Low()...)
	s.High().Error(err, tags.List(ts).All()...)
}

func (s *Adapter) Count(stat string, value float64, ts ...string) {
	s.CountH(stat, value, tags.LowList(ts...)...)
}

func (s *Adapter) Gauge(stat string, value float64, ts ...string) {
	s.GaugeH(stat, value, tags.LowList(ts...)...)
}

func (s *Adapter) Timing(stat string, value time.Duration, ts ...string) {
	s.TimingH(stat, value, tags.LowList(ts...)...)
}

func (s *Adapter) Histogram(stat string, value float64, ts ...string) {
	s.HistogramH(stat, value, tags.LowList(ts...)...)
}

func (s *Adapter) Unwrap() stats.HighCardinalityCollector {
	return s.HighCardinalityCollector
}

func (s *Adapter) With(ts ...string) *Adapter {
	return NewAdapter(s.HighCardinalityCollector.WithH(tags.LowList(ts...)...))
}

func (s *Adapter) WithH(tags ...tags.Tag) *Adapter {
	return NewAdapter(s.HighCardinalityCollector.WithH(tags...))
}
