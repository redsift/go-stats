package stats

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/PagerDuty/godspeed"
	"github.com/redsift/go-errs"
)

type dogstatsd struct {
	ns   string
	tags []string
	a    *godspeed.Godspeed
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

func send(a *godspeed.Godspeed, e *statsd) {
	if e.statsdDatum != nil {
		if err := a.Send(e.stat, e.kind, e.value, e.sampleRate, e.tags); err != nil {
			log.Printf("Unable to send stats data: %s", err)
		}
	} else {
		if err := a.Event(e.title, e.text, e.fields, e.tags); err != nil {
			log.Printf("Unable to send stats event: %s", err)
		}
	}
}

func NewDogstatsD(host string, port int, ns string, tags ...string) (Collector, error) {
	a, err := godspeed.New(host, port, false)
	if err != nil {
		return nil, fmt.Errorf("failed to created statsd client: %w", err)
	}

	a.SetNamespace(ns)
	a.AddTags(tags)

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
	fields := make(map[string]string)

	switch level {
	case Success:
		fields["alert_type"] = "success"
	case Warning:
		fields["alert_type"] = "warning"
	case Error:
		fields["alert_type"] = "error"
	case Info:
		fields["alert_type"] = "info"
	}

	if aggregation != "" {
		fields["aggregation_key"] = aggregation
	}

	if source != "" {
		fields["source_type_name"] = source
	}

	if low {
		fields["priority"] = "low"
	}

	if len(tags) == 0 {
		tags = []string{d.ns}
	} else {
		tags = append(tags, d.ns)
	}

	e := &statsd{nil, &statsdEvent{title, text, fields}, tags}
	send(d.a, e)
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
	d.a.Count(stat, count, tags)
}

func (d *dogstatsd) Gauge(stat string, value float64, tags ...string) {
	d.a.Gauge(stat, value, tags)
}

func (d *dogstatsd) Timing(stat string, value time.Duration, tags ...string) {
	d.a.Timing(stat, float64(value)/float64(time.Millisecond), tags)
}

func (d *dogstatsd) Histogram(stat string, value float64, tags ...string) {
	d.a.Histogram(stat, value, tags)
}

func (d *dogstatsd) Close() {
	// nothing to do!
	d.a.Conn.Close()
}

func (d *dogstatsd) Tags() []string {
	return d.tags
}

func (d *dogstatsd) With(tags ...string) Collector {
	return NewWithCollector(d, tags...)
}
