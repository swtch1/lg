package loadgen

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/swtch1/lg/store"
)

const scalePct = 10

func NewCoordinator(r LatencyReader, s FactorSetter, log *logrus.Entry) *Coordinator {
	return &Coordinator{
		read:        r,
		factorStore: s,
		log:         log,
	}
}

type (
	Coordinator struct {
		read        LatencyReader
		factorStore FactorSetter
		log         *logrus.Entry
	}

	LatencyReader interface {
		GetLatency() ([]store.AggLatency, error)
	}

	FactorSetter interface {
		SetScaleFactor(f float64) error
	}
)

func (c *Coordinator) Run(ctx context.Context, targetLatency int, measureTick time.Duration) error {
	scaleBy := float64(scalePct) / 100.0
	tgt := float64(targetLatency)

	var upperBound, lowerBound float64 = 3 * tgt, 0
	factor := upperBound
	if factor > 5000 {
		// don't let the scale factor start too high
		factor = float64(5000)
	}

	// the the factor initially
	err := c.factorStore.SetScaleFactor(factor)
	if err != nil {
		return fmt.Errorf("failed to set scale factor: %w", err)
	}

	ticker := time.NewTicker(measureTick)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			// TODO: we're just getting all latencies here but a better solution would be to get the
			// TODO: last n latencies so that we aren't grabbing every record every time
			ls, err := c.read.GetLatency()
			if err != nil {
				return fmt.Errorf("cannot get latency details from DB, abort: %w", err)
			}

			// TODO: likely best to keep an offset of records we've not seen, or delete processed records with a transaction
			// break this into a subset
			ls = lastNLatencies(ls, 10)

			// get the average
			current := avgLatencies(ls)

			if current < tgt {
				// we need to push harder
				upperBound = current

				// remember, since the factor here is a wait time we go down to speed up
				factor = lowerFactor(lowerBound, upperBound, scaleBy)
			} else {
				// we need to back off
				lowerBound = current

				// remember, since the factor here is a wait time we go up to slow down
				factor = raiseFactor(lowerBound, upperBound, scaleBy)
			}

			c.log.Trace("setting scale factor to %f", factor)
			err = c.factorStore.SetScaleFactor(factor)
			if err != nil {
				return fmt.Errorf("failed to set scale factor: %w", err)
			}
		}
	}
}

func lastNLatencies(ls []store.AggLatency, n int) []store.AggLatency {
	if len(ls) <= n {
		return ls
	}
	return ls[len(ls)-n:]
}

func avgLatencies(ls []store.AggLatency) float64 {
	var agg uint64
	for _, l := range ls {
		agg += l.LatencyMS
	}
	return float64(agg) / float64(len(ls))
}

func raiseFactor(min, max, pct float64) float64 {

}

func lowerFactor(min, max, pct float64) float64 {

}
