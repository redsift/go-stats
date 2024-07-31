package victoriametrics

import (
	"errors"
	"strings"
	"time"

	"github.com/launchdarkly/go-jsonstream/v3/jwriter"
)

type Metric struct {
	Name string
	Tags map[string]string
	Data struct {
		Values     []float64
		Timestamps []time.Time
	}
}

var (
	ErrNoData            = errors.New("metric has no data")
	ErrDataCountMismatch = errors.New("must have the same number of timestamps and values")
)

func (l Metric) MarshalJSON() ([]byte, error) {
	if len(l.Data.Values) != len(l.Data.Timestamps) {
		return nil, ErrDataCountMismatch
	}

	if len(l.Data.Values) == 0 || len(l.Data.Timestamps) == 0 {
		return nil, ErrNoData
	}

	w := jwriter.NewWriter()
	o := w.Object()

	mo := o.Name("metric").Object()
	mo.Name("__name__").String(l.Name)
	for k, v := range l.Tags {
		mo.Name(k).String(v)
	}
	mo.End()

	va := o.Name("values").Array()
	for _, value := range l.Data.Values {
		va.Float64(value)
	}
	va.End()

	ta := o.Name("timestamps").Array()
	for _, t := range l.Data.Timestamps {
		ta.Int(int(t.UnixMilli()))
	}
	ta.End()

	o.End()

	w.Flush()
	return w.Bytes(), w.Error()
}

func (l *Metric) addLegacyTags(tags []string) {
	if l.Tags == nil {
		l.Tags = make(map[string]string)
	}

	for _, tag := range tags {
		k, v, ok := strings.Cut(tag, ":")
		if !ok {
			continue
		}
		l.Tags[k] = v
	}
}

func (l *Metric) AddValue(t time.Time, v float64) {
	l.Data.Timestamps = append(l.Data.Timestamps, t)
	l.Data.Values = append(l.Data.Values, v)
}
