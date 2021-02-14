package loadgen

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/swtch1/lg/store"
)

// scalePct should be a number between 0 and 1 which tells the coordinator how much to move the high or low water mark
const scalePct = 0.1

func NewCoordinator(ls LatencyReadPurger, fs FactorSetter, log *logrus.Entry) *Coordinator {
	return &Coordinator{
		latencyStore: ls,
		factorStore:  fs,
		log:          log,
	}
}

type (
	Coordinator struct {
		latencyStore LatencyReadPurger
		factorStore  FactorSetter
		log          *logrus.Entry
	}

	LatencyReadPurger interface {
		GetLatency() ([]store.AggLatency, error)
		PurgeLatencies() error
	}

	FactorSetter interface {
		SetScaleFactor(f float64) error
	}
)

func (c *Coordinator) Run(ctx context.Context, targetLatency int, measureTick time.Duration) error {
	tgt := float64(targetLatency)

	var upperBound, lowerBound float64 = 3 * tgt, 0
	factor := upperBound
	if factor > 5000 {
		// don't let the scale factor start too high
		factor = float64(5000)
	}

	err := c.latencyStore.PurgeLatencies()
	if err != nil {
		return fmt.Errorf("failed to purge latency records: %w", err)
	}

	// the the factor initially
	err = c.factorStore.SetScaleFactor(factor)
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
			ls, err := c.latencyStore.GetLatency()
			if err != nil {
				return fmt.Errorf("cannot get latency details from DB, abort: %w", err)
			}

			// TODO: likely best to keep an offset of records we've not seen, or delete processed records with a transaction
			// break this into a subset
			ls = lastNLatencies(ls, 10)

			// get the average
			current := avgLatencies(ls)

			if current < tgt {
				decreaseUpper(&upperBound, lowerBound, scalePct)
				factor = upperBound
			} else {
				increaseLower(&lowerBound, upperBound, scalePct)
				factor = lowerBound
			}

			c.log.WithFields(logrus.Fields{
				"latencyAvg": fmt.Sprintf("%.3f", current),
				"factor":     fmt.Sprintf("%.3f", factor),
				"boundLower": fmt.Sprintf("%.3f", lowerBound),
				"boundUpper": fmt.Sprintf("%.3f", upperBound),
			}).Trace("setting scale factor")
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
	var agg float64
	for _, l := range ls {
		agg += l.LatencyMS
	}
	return agg / float64(len(ls))
}

func increaseLower(lowerBound *float64, upperBound, pct float64) {
	diff := upperBound - *lowerBound
	*lowerBound = *lowerBound + (diff * pct)
}

func decreaseUpper(upperBound *float64, lowerBound, pct float64) {
	diff := *upperBound - lowerBound
	*upperBound = *upperBound - (diff * pct)
}
