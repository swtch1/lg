package loadgen

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/swtch1/lg/domain"
	"github.com/swtch1/lg/store"
)

func NewGenerator(targetAddr string, f Feeder, w LatencyWriter, g FactorGetter, log *logrus.Entry) *Generator {
	return &Generator{
		info:        newNodeInfo(),
		baseAddr:    targetAddr,
		dispatcher:  newDispatcher(time.Second * 3),
		feed:        f,
		write:       w,
		factorStore: g,
		log:         log,
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
		baseAddr    string
		dispatcher  *dispatcher
		feed        Feeder
		write       LatencyWriter
		factorStore FactorGetter
		log         *logrus.Entry
	}

	// Feeder gives the next RRPair to be processed.
	Feeder interface {
		Next() (domain.RRPair, error)
	}

	LatencyWriter interface {
		CreateLatencies(ls []store.AggLatency) error
	}
	FactorGetter interface {
		GetScaleFactor() (float64, error)
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
			if err := g.pause(); err != nil {
				g.log.WithError(err).Error("pause failed and we cannot proceed reliably without it, skipping call")
				continue
			}
			g.feedAndReport()
		}
	}
}

// pause so the SUT is not overloaded.
func (g *Generator) pause() error {
	// TODO: the scale factor is retrieved for every single requests here which likely doesn't make sense
	// TODO: instead we should retrieve this every so often and serve the same number for multiple requests
	f, err := g.factorStore.GetScaleFactor()
	if err != nil {
		return fmt.Errorf("failed to get scale factor: %w", err)
	}
	time.Sleep(time.Millisecond * time.Duration(f))
	return nil
}

func (g *Generator) feedAndReport() {
	pair, err := g.feed.Next()
	if err != nil {
		g.log.WithError(err).Error("failed to get next RR pair")
		return
	}
	pair.Req.Path = strings.TrimRight(g.baseAddr, "/") + "/" + pair.Req.Path

	resp, m, err := g.dispatcher.dispatch(pair.Req, g.info, g.log)
	if err != nil {
		g.log.WithError(err).Error("failed to make successful outbound request")
		return
	}
	if !g.respMatch(pair.Resp, resp) {
		// TODO: we need to do something more with this than just logging it
		// TODO: should be taken into account in a broader context
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
