package stats

import (
	"errors"
	"fmt"
	"log"
	"time"

	ddsd "github.com/DataDog/datadog-go/v5/statsd"
	"github.com/redsift/go-errs"
)

const sendBuffer = 16

type dogstatsd struct {
	ns   string
	tags []string
	a    *ddsd.Client
}

type statsd struct {
	*statsdDatum
	*statsdEvent

	tags []string
}

type statsdDatum struct {
	stat       string
	kind       string
	value      float64
	sampleRate float64
}

type statsdEvent struct {
	title  string
	text   string
	fields map[string]string
}

func NewDogstatsD(host string, port int, ns string, tags ...string) (Collector, error) {
	a, err := ddsd.New(fmt.Sprintf("%s:%d", host, port), ddsd.WithNamespace(ns), ddsd.WithTags(tags))
	if err != nil {
		return nil, fmt.Errorf("failed to created statsd client: %w", err)
	}

	return &dogstatsd{ns, tags, a}, nil
}

type EventLevel int

const (
	Info EventLevel = iota
	Success
	Warning
	Error
)

// Use aggregation as a key to group events together.
// Events are aggregated on the Event Stream based on: hostname/level/source/aggregation.
// Use source string to identify the source of the event.
// Set low to true if the event has a low priority.
func (d *dogstatsd) event(level EventLevel, title, text, source, aggregation string, low bool, tags ...string) {
	ev := ddsd.NewEvent(title, text)
	switch level {
	case Success:
		ev.AlertType = ddsd.Success
	case Warning:
		ev.AlertType = ddsd.Warning
	case Error:
		ev.AlertType = ddsd.Error
	case Info:
		ev.AlertType = ddsd.Info
	}

	if aggregation != "" {
		ev.AggregationKey = aggregation
	}

	if source != "" {
		ev.SourceTypeName = source
	}

	if low {
		ev.Priority = ddsd.Low
	}

	if len(tags) == 0 {
		tags = []string{d.ns}
	} else {
		tags = append(tags, d.ns)
	}

	ev.Tags = tags

	d.a.Event(ev)
}

func (d *dogstatsd) Inform(title, text string, tags ...string) {
	d.event(Info, title, text, "", "", true, tags...)
}

func (d *dogstatsd) Error(err error, tags ...string) {
	if err == nil {
		return
	}

	var pe *errs.PropagatedError
	if !errors.As(err, &pe) {
		return
	}

	title := pe.Code.String() + " / " + pe.Title + " / " + pe.Id
	text := pe.Detail + " / " + pe.Link
	src := ""

	if pe.Source != nil {
		if src = pe.Source.Parameter; src == "" {
			src = "jsonpointer:" + pe.Source.Pointer
		}
	}

	agg := pe.Code.String()
	d.event(Error, title, text, src, agg, false, tags...)
}

func (d *dogstatsd) Count(stat string, count float64, tags ...string) {
	if err := d.a.Count(stat, int64(count), tags, 1); err != nil {
		log.Printf("Unable to send stats data: %s", err)
	}
}

func (d *dogstatsd) Gauge(stat string, value float64, tags ...string) {
	if err := d.a.Gauge(stat, value, tags, 1); err != nil {
		log.Printf("Unable to send stats data: %s", err)
	}
}

func (d *dogstatsd) Timing(stat string, value time.Duration, tags ...string) {
	if err := d.a.Timing(stat, value, tags, 1); err != nil {
		log.Printf("Unable to send stats data: %s", err)
	}
}

func (d *dogstatsd) Histogram(stat string, value float64, tags ...string) {
	if err := d.a.Histogram(stat, value, tags, 1); err != nil {
		log.Printf("Unable to send stats data: %s", err)
	}
}

func (d *dogstatsd) Close() {
	d.a.Close()
}

func (d *dogstatsd) Tags() []string {
	return d.tags
}

func (d *dogstatsd) With(tags ...string) Collector {
	return NewWithCollector(d, tags...)
}
