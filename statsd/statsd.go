//
// Copyright (c) 2016-2021 Redsift Limited. All rights reserved.
//
package statsd

import (
	env "github.com/redsift/go-cfg"
	"github.com/redsift/go-stats/stats"
)

const (
	EnvNamespace = "STATSD_NAMESPACE"
	EnvHost      = "STATSD_HOST"
	EnvPort      = "STATSD_PORT"
	EnvTags      = "STATSD_TAGS"
)

type config struct {
	namespace string
	host      string
	port      int
	tags      []string
}

type Option func(*config)

func OptHost(s string) Option      { return func(c *config) { c.host = s } }
func OptPort(n int) Option         { return func(c *config) { c.port = n } }
func OptNamespace(s string) Option { return func(c *config) { c.namespace = s } }
func OptTags(l ...string) Option   { return func(c *config) { c.tags = l } }

func New(opts ...Option) stats.Collector {
	cfg := &config{
		host:      env.EnvString(EnvHost, "127.0.0.1"),
		port:      env.EnvInt(EnvPort, 8125),
		namespace: env.EnvString(EnvNamespace, ""),
		tags:      env.EnvStringArray(EnvTags),
	}

	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.namespace == "" {
		return stats.NewDiscardCollector()
	}

	collector, err := stats.NewDogstatsD(cfg.host, cfg.port, cfg.namespace, cfg.tags...)
	if err != nil {
		return stats.NewDiscardCollector()
	}

	return collector
}

type BeanspikeStats struct {
	collector stats.Collector
}

func (bsStats BeanspikeStats) Handler(event, tube string, count float64) {
	bsStats.collector.Count(event, count, []string{tube}...)
}

func NewBeanspikeStats(collector stats.Collector) BeanspikeStats {
	return BeanspikeStats{collector: collector}
}
