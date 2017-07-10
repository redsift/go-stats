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

	"github.com/redsift/go-cfg"
	"github.com/redsift/go-stats/stats"
)

const defaultPort = 8125

var namespace string
var host string
var port int
var gtags []string

var parsed bool

func Init() {
	namespace = cfg.EnvString("STATSD_NAMESPACE", "")
	host = cfg.EnvString("STATSD_HOST", "127.0.0.1")
	port = cfg.EnvInt("STATSD_PORT", defaultPort)
	gtags = cfg.EnvStringArray("STATSD_TAGS")

	parsed = true
}

func New() stats.Collector {
	if !parsed {
		Init()
	}

	if namespace == "" {
		fmt.Println("No stats collector specified, sinking to null")
		return stats.NewNull()
	}

	tags := gtags

	var err error
	collector, err := stats.NewDogstatsD(host, port, namespace, tags)
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
	bsStats.collector.Count(event, count, []string{tube})
}

func NewBeanspikeStats(collector stats.Collector) BeanspikeStats {
	return BeanspikeStats{collector: collector}
}
