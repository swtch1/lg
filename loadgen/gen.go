package loadgen

import (
	"bytes"
	"context"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/swtch1/lg/domain"
	"github.com/swtch1/lg/store"
)

func NewGenerator(targetAddr string, f Feeder, w LatencyWriter, log *logrus.Entry) *Generator {
	return &Generator{
		info:       newNodeInfo(),
		baseAddr:   targetAddr,
		dispatcher: newDispatcher(time.Second * 3),
		feed:       f,
		write:      w,
		log:        log,
	}
}

type (
	// Generator should:
	//  - create load
	//  - throw load at the SUT
	//  - store and report aggregate metrics
	Generator struct {
		info nodeInfo
		// baseAddr is the base address where load should be sent
		baseAddr   string
		dispatcher *dispatcher
		feed       Feeder
		write      LatencyWriter
		log        *logrus.Entry
	}

	// Feeder gives the next RRPair to be processed.
	Feeder interface {
		Next() (domain.RRPair, error)
	}

	LatencyWriter interface {
		CreateLatencies(ls []store.AggLatency) error
	}
)

// Run the load generator until a fatal error occurs.
func (g *Generator) Run(ctx context.Context, goroutines int) {
	var wg sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go g.loadEmUp(ctx, &wg)
	}
	wg.Wait()
}

func (g *Generator) loadEmUp(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			g.feedAndReport()
		}
	}
}

func (g *Generator) feedAndReport() {
	pair, err := g.feed.Next()
	if err != nil {
		g.log.WithError(err).Error("failed to get next RR pair")
		return
	}
	resp, m, err := g.dispatcher.dispatch(pair.Req, g.info, g.log)
	if err != nil {
		g.log.WithError(err).Error("failed to make successful outbound request")
		return
	}
	if !g.respMatch(pair.Resp, resp) {
		// TODO: we need to do something more with this than just logging it
		// TODO: should be taken into account in a braoder context
		g.log.Error("response did not match")
		// don't log latency for failed calls
		return
	}

	err = g.write.CreateLatencies([]store.AggLatency{m.latency})
	if err != nil {
		g.log.WithError(err).Error("failed to store latency.. this could be a problem")
		return
	}
}

func (g *Generator) respMatch(r1, r2 domain.Response) bool {
	if bytes.Compare(r1.Body, r2.Body) != 0 {
		return false
	}
	if r1.StatusCode != r2.StatusCode {
		return false
	}
	return true
}
