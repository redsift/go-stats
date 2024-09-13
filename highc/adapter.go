package highc

import (
	"time"

	"github.com/redsift/go-stats/stats"
)

func NewAdapter(hc stats.HighCardinalityCollector) *Adapter {
	return &Adapter{
		HighCardinalityCollector: hc,
	}
}

type Adapter struct {
	stats.HighCardinalityCollector
}

func (s *Adapter) Inform(title, text string, tags ...string) {
	s.Low().Inform(title, text, tags...)
	s.High().Inform(title, text, tags...)
}

func (s *Adapter) Error(err error, tags ...string) {
	s.Low().Error(err, tags...)
	s.High().Error(err, tags...)
}

func (s *Adapter) Count(stat string, value float64, tags ...string) {
	s.CountH(stat, value, tags, nil)
}

func (s *Adapter) Gauge(stat string, value float64, tags ...string) {
	s.GaugeH(stat, value, tags, nil)
}

func (s *Adapter) Timing(stat string, value time.Duration, tags ...string) {
	s.TimingH(stat, value, tags, nil)
}

func (s *Adapter) Histogram(stat string, value float64, tags ...string) {
	s.HistogramH(stat, value, tags, nil)
}

func (s *Adapter) Unwrap() stats.HighCardinalityCollector {
	return s.HighCardinalityCollector
}

func (s *Adapter) With(tags ...string) *Adapter {
	return NewAdapter(s.HighCardinalityCollector.WithH(tags, nil))
}

func (s *Adapter) WithH(low, high []string) *Adapter {
	return NewAdapter(s.HighCardinalityCollector.WithH(low, high))
}
