//
// statsd.go
//
//
// Copyright (c) 2016 Redsift Limited. All rights reserved.
//

package statsd

import (
	"fmt"
	"os"
	"sync"

	"github.com/redsift/go-cfg"
	"github.com/redsift/go-stats/stats"
)

const defaultPort = 8125

var (
	namespace string
	host      string
	port      int
	gtags     []string
	once      sync.Once
)

func configure() {
	namespace = cfg.EnvString("STATSD_NAMESPACE", "")
	host = cfg.EnvString("STATSD_HOST", "127.0.0.1")
	port = cfg.EnvInt("STATSD_PORT", defaultPort)
	gtags = cfg.EnvStringArray("STATSD_TAGS")
}

func Init() {
	once.Do(configure)
}

func New() stats.Collector {
	once.Do(configure)

	if namespace == "" {
		fmt.Println("No stats collector specified, sinking to null")
		return stats.NewDiscardCollector()
	}

	tags := gtags

	var err error
	collector, err := stats.NewDogstatsD(host, port, namespace, tags...)
	if err != nil {
		fmt.Printf("Could not create DogstatsD collector: %s\n", err)
		os.Exit(1)
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
