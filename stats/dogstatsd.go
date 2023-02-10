package stats

import (
	"errors"
	"fmt"
	"log"
	"time"

	datadog "github.com/DataDog/datadog-go/v5/statsd"
	"github.com/redsift/go-errs"
)

type dogstatsd struct {
	ns string
	a  *datadog.Client
}

func NewDogstatsD(host string, port int, ns string, tags ...string) (Collector, error) {
	a, err := datadog.New(fmt.Sprintf("%s:%d", host, port), datadog.WithNamespace(ns))
	if err != nil {
		return nil, fmt.Errorf("failed to create statsd client: %w", err)
	}

	return (&dogstatsd{ns, a}).With(tags...), nil
}

func NewWithDogClient(client *datadog.Client, ns string, tags ...string) (Collector, error) {
	a, err := datadog.CloneWithExtraOptions(client, datadog.WithNamespace(ns))
	if err != nil {
		return nil, fmt.Errorf("failed to create statsd client: %w", err)
	}

	return (&dogstatsd{ns, a}).With(tags...), nil
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
	ev := datadog.NewEvent(title, text)
	switch level {
	case Success:
		ev.AlertType = datadog.Success
	case Warning:
		ev.AlertType = datadog.Warning
	case Error:
		ev.AlertType = datadog.Error
	case Info:
		ev.AlertType = datadog.Info
	}

	if aggregation != "" {
		ev.AggregationKey = aggregation
	}

	if source != "" {
		ev.SourceTypeName = source
	}

	if low {
		ev.Priority = datadog.Low
	}

	if len(tags) == 0 {
		tags = []string{d.ns}
	} else {
		tags = append(tags, d.ns)
	}

	ev.Tags = tags

	if err := d.a.Event(ev); err != nil {
		log.Printf("Unable to send stats event: %s", err)
	}
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
	return nil
}

func (d *dogstatsd) Parent() Collector {
	return nil
}

func (d *dogstatsd) With(tags ...string) Collector {
	return NewWithCollector(d, tags...)
}

func (d *dogstatsd) WithoutTags() Collector {
	return d
}
