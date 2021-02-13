package mysql

import (
	"github.com/swtch1/lg/store"
)

func NewCollector(ls LatencyStore) *Collector {
	return &Collector{
		ls: ls,
	}
}

// Collector sends
type (
	Collector struct {
		ls LatencyStore
	}

	LatencyStore interface {
		CreateLatencies(ls []store.AggLatency) error
	}
)

func (c *Collector) WriteLatency(l store.AggLatency) error {
	return c.ls.CreateLatencies([]store.AggLatency{l})
}
