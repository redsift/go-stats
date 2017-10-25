package stats

import (
	"strings"
	"time"

	"golang.org/x/net/context"
)

// Tags should use a key:value format

type Collector interface {
	// FYI entry
	Inform(title, text string, tags ...string)

	// Error resulting in a notification
	Error(e error, tags ...string)

	// Measure rate of events over dT, an Inc = Count(1), Dec = Count(-1)
	Count(stat string, count float64, tags ...string)

	// Log the value at T
	Gauge(stat string, value float64, tags ...string)

	// Log the value at T for count/avg/median/max/95percentile
	Timing(stat string, value time.Duration, tags ...string)

	Histogram(stat string, value float64, tags ...string)

	Close()

	Tags() []string
}

const (
	ctxKeyCollector = "redsift/go-utils/stats#Collector"
)

// ContextWithCollector creates a new context with an instance of Collector
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

func NewDiscardCollector() Collector                                      { return &discardCollector{} }
func (_ *discardCollector) Inform(_, _ string, _ ...string)               {}
func (_ *discardCollector) Error(_ error, _ ...string)                    {}
func (_ *discardCollector) Count(_ string, _ float64, _ ...string)        {}
func (_ *discardCollector) Gauge(_ string, _ float64, _ ...string)        {}
func (_ *discardCollector) Timing(_ string, _ time.Duration, _ ...string) {}
func (_ *discardCollector) Histogram(_ string, _ float64, _ ...string)    {}
func (_ *discardCollector) Close()                                        {}
func (_ *discardCollector) Tags() []string                                { return nil }


type wrap struct {
		c    Collector
		tags []string
	}

func Wrap(c Collector, tags []string) Collector { return &wrap{c, tags} }

func (w *wrap) Inform(title, text string, tags ...string) {
		w.c.Inform(title, text, append(tags, w.tags...)...)
	}

func (w *wrap) Error(err error, tags ...string) {
		w.c.Error(err, append(tags, w.tags...)...)
	}

func (w *wrap) Count(stat string, count float64, tags ...string) {
		w.c.Count(stat, count, append(tags, w.tags...)...)
	}

func (w *wrap) Gauge(stat string, value float64, tags ...string) {
		w.c.Gauge(stat, value, append(tags, w.tags...)...)
	}

func (w *wrap) Timing(stat string, value time.Duration, tags ...string) {
		w.c.Timing(stat, value, append(tags, w.tags...)...)
	}

func (w *wrap) Histogram(stat string, value float64, tags ...string) {
		w.c.Histogram(stat, value, append(tags, w.tags...)...)
	}

func (w *wrap) Close() { w.c.Close() }

func (w *wrap) Tags() []string { return w.c.Tags() }

// Safe for concurrent use
var replacer = strings.NewReplacer(" ", "_", ".", "_")

// lowercase, no '. '
func Sanitise(in string) string { return replacer.Replace(strings.ToLower(in)) }
