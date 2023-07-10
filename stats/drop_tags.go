package stats

import (
	"time"

	"github.com/samber/lo"
)

type DropTags struct {
	next   Collector
	remove []string
}

func NewTagDropper(collector Collector, tags ...string) Collector {
	return &DropTags{
		next:   collector,
		remove: tags,
	}
}

func (w *DropTags) Inform(title, text string, tags ...string) {
	w.next.Inform(title, text, w.filterTags(tags)...)
}

func (w *DropTags) Error(err error, tags ...string) {
	w.next.Error(err, w.filterTags(tags)...)
}

func (w *DropTags) Count(name string, value float64, tags ...string) {
	w.next.Count(name, value, w.filterTags(tags)...)
}

func (w *DropTags) Gauge(name string, value float64, tags ...string) {
	w.next.Gauge(name, value, w.filterTags(tags)...)
}

func (w *DropTags) Timing(name string, value time.Duration, tags ...string) {
	w.next.Timing(name, value, w.filterTags(tags)...)
}

func (w *DropTags) Histogram(name string, value float64, tags ...string) {
	w.next.Histogram(name, value, w.filterTags(tags)...)
}

func (w *DropTags) Close() {
	w.next.Close()
}

func (w *DropTags) With(tags ...string) Collector {
	return NewTagDropper(w.next.With(tags...), w.remove...)
}

func (w *DropTags) Tags() []string {
	return w.filterTags(w.next.Tags())
}

func (w *DropTags) filterTags(tags []string) []string {
	return lo.Without(tags, w.remove...)
}
