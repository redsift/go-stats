package wsendpoint

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/redsift/go-errs"
	"github.com/redsift/go-stats/stats"
	"gopkg.in/launchdarkly/go-jsonstream.v1/jwriter"
)

type endpoint struct {
	mux    chan []byte
	reg    chan muxreg
	cancel context.CancelFunc
}

type muxreg struct {
	ctx context.Context
	ch  chan<- []byte
}

func New(ctx context.Context) (http.Handler, stats.Collector) {
	ctx, cancel := context.WithCancel(ctx)
	e := &endpoint{
		mux:    make(chan []byte),
		reg:    make(chan muxreg),
		cancel: cancel,
	}
	go e.run(ctx)
	return e, e
}

func (e *endpoint) run(ctx context.Context) {
	var o []muxreg
	for {
		select {
		case <-ctx.Done():
			return
		case i := <-e.mux:
			oo := make([]muxreg, 0, len(o))
			for _, reg := range o {
				if reg.ctx.Err() != nil {
					close(reg.ch)
					continue
				}
				oo = append(oo, reg)
				reg.ch <- i
			}
			o = oo
		case muxreg := <-e.reg:
			o = append(o, muxreg)
		}
	}
}

func (e *endpoint) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sock, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		w.WriteHeader(http.StatusUpgradeRequired)
		return
	}
	defer sock.Close()

	ch := make(chan []byte)

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	reg := muxreg{ctx, ch}
	e.reg <- reg

	for item := range ch {
		if err := wsutil.WriteServerText(sock, item); err != nil {
			return
		}
	}
}

type EventLevel int

const (
	Info EventLevel = iota
	Success
	Warning
	Error
)

func (e *endpoint) writeObj(name string, fn func(jwriter.ObjectState)) {
	b := &bytes.Buffer{}
	w := jwriter.NewStreamingWriter(b, 1024)
	oo := w.Object()
	o := oo.Name(name).Object()
	fn(o)
	o.End()
	oo.End()
	w.Flush()
	e.mux <- b.Bytes()
}

func (d *endpoint) event(level EventLevel, title, text, source, aggregation string, low bool, tags ...string) {
	d.writeObj("event", func(o jwriter.ObjectState) {
		switch level {
		case Success:
			o.Name("level").String("success")
		case Warning:
			o.Name("level").String("warning")
		case Error:
			o.Name("level").String("error")
		case Info:
			o.Name("level").String("info")
		}
		o.Maybe("title", title != "").String(title)
		o.Maybe("text", text != "").String(text)
		o.Maybe("source", source != "").String(source)
		o.Maybe("aggregation", aggregation != "").String(aggregation)
		tagObject(o, tags)
	})
}

func tagObject(o jwriter.ObjectState, tags []string) {
	if len(tags) == 0 {
		return
	}

	tp := o.Name("tags").Object()
	for _, t := range tags {
		k, v, _ := strings.Cut(t, ":")
		tp.Name(k).String(v)
	}
	tp.End()
}

func (d *endpoint) Inform(title, text string, tags ...string) {
	d.event(Info, title, text, "", "", true, tags...)
}

func (d *endpoint) Error(err error, tags ...string) {
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

func (d *endpoint) Count(stat string, count float64, tags ...string) {
	d.writeObj("count", func(o jwriter.ObjectState) {
		o.Name("name").String(stat)
		o.Name("count").Float64(count)
		tagObject(o, tags)
	})
}

func (d *endpoint) Gauge(stat string, value float64, tags ...string) {
	d.writeObj("gauge", func(o jwriter.ObjectState) {
		o.Name("name").String(stat)
		o.Name("count").Float64(value)
		tagObject(o, tags)
	})
}

func (d *endpoint) Timing(stat string, value time.Duration, tags ...string) {
	d.writeObj("timing", func(o jwriter.ObjectState) {
		o.Name("name").String(stat)
		o.Name("count").String(value.String())
		tagObject(o, tags)
	})
}

func (d *endpoint) Histogram(stat string, value float64, tags ...string) {
	d.writeObj("histogram", func(o jwriter.ObjectState) {
		o.Name("name").String(stat)
		o.Name("count").Float64(value)
		tagObject(o, tags)
	})
}

func (e *endpoint) Close() {
	e.cancel()
}

func (e *endpoint) Tags() []string {
	return nil
}

func (e *endpoint) With(tags ...string) stats.Collector {
	return stats.NewWithCollector(e, tags...)
}
