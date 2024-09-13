package highc

import "github.com/redsift/go-stats/stats"

func NewDiscard() stats.HighCardinalityCollector {
	return New(
		stats.NewDiscardCollector(),
		stats.NewDiscardCollector(),
	)
}
