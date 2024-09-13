package router

import (
	"time"

	"github.com/redsift/go-stats/stats"
)

type Rule interface {
	Match(name string, tags []string) bool
	Not() Rule
}

type Route struct {
	Rule
	Target stats.Collector
}

type router struct {
	routes []Route
}

func To(target stats.Collector, rule Rule) Route {
	return Route{
		Target: target,
		Rule:   rule,
	}
}

func New(routes ...Route) stats.Collector {
	if len(routes) == 0 {
		return stats.NewDiscardCollector()
	}
	return &router{routes: routes}
}

func (s *router) Inform(title, text string, tags ...string) {
	for _, route := range s.routes {
		if route.Match(title, tags) {
			route.Target.Inform(title, text, tags...)
		}
	}
}

func (s *router) Error(err error, tags ...string) {
	e := err.Error()
	for _, rule := range s.routes {
		if rule.Match(e, tags) {
			rule.Target.Error(err, tags...)
		}
	}
}

func (s *router) Count(stat string, value float64, tags ...string) {
	for _, rule := range s.routes {
		if rule.Match(stat, tags) {
			rule.Target.Count(stat, value, tags...)
		}
	}
}

func (s *router) Gauge(stat string, value float64, tags ...string) {
	for _, rule := range s.routes {
		if rule.Match(stat, tags) {
			rule.Target.Gauge(stat, value, tags...)
		}
	}
}

func (s *router) Timing(stat string, value time.Duration, tags ...string) {
	for _, rule := range s.routes {
		if rule.Match(stat, tags) {
			rule.Target.Timing(stat, value, tags...)
		}
	}
}

func (s *router) Histogram(stat string, value float64, tags ...string) {
	for _, rule := range s.routes {
		if rule.Match(stat, tags) {
			rule.Target.Histogram(stat, value, tags...)
		}
	}
}

func (s *router) Close() {
	for _, rule := range s.routes {
		rule.Target.Close()
	}
}

func (s *router) Tags() []string {
	return nil
}

func (s *router) With(tags ...string) stats.Collector {
	return stats.NewWithCollector(s, tags...)
}

func (s *router) Unwrap() stats.Collector {
	r := &router{}
	for _, route := range s.routes {
		r.routes = append(r.routes, Route{
			Rule:   route.Rule,
			Target: stats.Unwrap(route.Target),
		})
	}
	return r
}
