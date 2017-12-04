package stats

import (
	"log"
	"time"

	"github.com/redsift/go-errs"

	"github.com/PagerDuty/godspeed"
)

const sendBuffer = 16

type dogstatsd struct {
	send chan *statsd
	ns   string
	tags []string
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
	a, err := godspeed.New(host, port, false)
	if err != nil {
		return nil, err
	}

	a.SetNamespace(ns)
	a.AddTags(tags)

	ch := make(chan *statsd, sendBuffer)

	go func() {
		for e := range ch {
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
	}()

	return &dogstatsd{ch, ns, tags}, nil
}

type EventLevel int

const (
	Info EventLevel = iota
	Success
	Warning
	Error
)

// aggregation : key to group events together. Events are aggregated on the Event Stream based on: hostname/level/source/aggregation
// source : string to identify source
// low : Low priority event
func (d *dogstatsd) event(level EventLevel, title, text, source, aggregation string, low bool, tags ...string) {
	fields := make(map[string]string)

	switch level {
	case Success:
		fields["alert_type"] = "success"
	case Warning:
		fields["alert_type"] = "warning"
	case Error:
		fields["alert_type"] = "error"
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
	d.send <- &statsd{nil, &statsdEvent{title, text, fields}, tags}
}

func (d *dogstatsd) Inform(title, text string, tags ...string) {
	d.event(Info, title, text, "", "", true, tags...)
}

func (d *dogstatsd) Error(err error, tags ...string) {
	if err == nil {
		return
	}
	pe, ok := err.(*errs.PropagatedError)
	if !ok {
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
	d.send <- &statsd{&statsdDatum{stat, "c", count, 1}, nil, tags}
}

func (d *dogstatsd) Gauge(stat string, value float64, tags ...string) {
	d.send <- &statsd{&statsdDatum{stat, "g", value, 1}, nil, tags}
}

func (d *dogstatsd) Timing(stat string, value time.Duration, tags ...string) {
	d.send <- &statsd{&statsdDatum{stat, "ms", float64(value) / float64(time.Millisecond), 1}, nil, tags}
}

func (d *dogstatsd) Histogram(stat string, value float64, tags ...string) {
	d.send <- &statsd{&statsdDatum{stat, "h", value, 1}, nil, tags}
}

func (d *dogstatsd) Close() {
	close(d.send)
}

func (d *dogstatsd) Tags() []string {
	return d.tags
}

func (d *dogstatsd) With(tags ...string) Collector {
	return newWithCollector(d, tags...)
}
