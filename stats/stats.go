package stats

import (
	"context"
	"strings"
	"time"
)

// Tags should use a key:value format

type Collector interface {
	// FYI entry
	Inform(string, string, ...string)

	// Error resulting in a notification
	Error(error, ...string)

	// Measure rate of events over dT, an Inc = Count(1), Dec = Count(-1)
	Count(string, float64, ...string)

	// Log the value at T
	Gauge(string, float64, ...string)

	// Log the value at T for count/avg/median/max/95percentile
	Timing(string, time.Duration, ...string)

	Histogram(string, float64, ...string)

	Close()

	Tags() []string

	With(...string) Collector
}

type HighCardinalityCollector interface {
	CountH(name string, value float64, low, high []string)
	GaugeH(name string, value float64, low, high []string)
	TimingH(name string, value time.Duration, low, high []string)
	HistogramH(name string, value float64, low, high []string)
	WithH(low, high []string) HighCardinalityCollector

	High() Collector
	Low() Collector
}

type contextKey struct {
	v string
}

//nolint:gochecknoglobals
var ctxKeyCollector = contextKey{"collector"}

// ContextWithCollector creates a new context with an instance of Collector.
func ContextWithCollector(ctx context.Context, c Collector) context.Context {
	return context.WithValue(ctx, ctxKeyCollector, c)
}

// CollectorFromContext returns Collector from ctx.
// The function returns nil if no Collector was stored in the Context.
func CollectorFromContext(ctx context.Context) Collector {
	if c := ctx.Value(ctxKeyCollector); c != nil {
		return c.(Collector)
	}

	return nil
}

type discardCollector struct{}

func NewDiscardCollector() Collector                                    { return &discardCollector{} }
func (*discardCollector) Inform(_, _ string, _ ...string)               {}
func (*discardCollector) Error(_ error, _ ...string)                    {}
func (*discardCollector) Count(_ string, _ float64, _ ...string)        {}
func (*discardCollector) Gauge(_ string, _ float64, _ ...string)        {}
func (*discardCollector) Timing(_ string, _ time.Duration, _ ...string) {}
func (*discardCollector) Histogram(_ string, _ float64, _ ...string)    {}
func (*discardCollector) Close()                                        {}
func (*discardCollector) Tags() []string                                { return nil }
func (dc *discardCollector) With(...string) Collector                   { return dc }

// Safe for concurrent use.
//
//nolint:gochecknoglobals
var replacer = strings.NewReplacer(" ", "_", ".", "_")

// Sanitise returns copy of in string in lowercase; ' ' and '.' replaced with '_'.
func Sanitise(in string) string { return replacer.Replace(strings.ToLower(in)) }

type withCollector struct {
	tags []string
	c    Collector
}

func NewWithCollector(c Collector, tags ...string) Collector {
	return &withCollector{
		tags: tags,
		c:    c,
	}
}

func (wc *withCollector) allTags(tags ...string) []string {
	t := make([]string, 0, len(wc.tags)+len(tags))

	return append(append(t, wc.tags...), tags...)
}

func (wc *withCollector) Inform(title, text string, tags ...string) {
	wc.c.Inform(title, text, wc.allTags(tags...)...)
}

func (wc *withCollector) Error(err error, tags ...string) {
	wc.c.Error(err, wc.allTags(tags...)...)
}

func (wc *withCollector) Count(stat string, count float64, tags ...string) {
	wc.c.Count(stat, count, wc.allTags(tags...)...)
}

func (wc *withCollector) Gauge(stat string, value float64, tags ...string) {
	wc.c.Gauge(stat, value, wc.allTags(tags...)...)
}

func (wc *withCollector) Timing(stat string, value time.Duration, tags ...string) {
	wc.c.Timing(stat, value, wc.allTags(tags...)...)
}

func (wc *withCollector) Histogram(stat string, value float64, tags ...string) {
	wc.c.Histogram(stat, value, wc.allTags(tags...)...)
}

func (wc *withCollector) Close() {
	wc.c.Close()
}

func (wc *withCollector) Tags() []string {
	ot := wc.c.Tags()
	t := make([]string, 0, len(wc.tags)+len(ot))

	return append(append(t, ot...), wc.tags...)
}

func (wc *withCollector) With(tags ...string) Collector {
	return NewWithCollector(wc, tags...)
}
