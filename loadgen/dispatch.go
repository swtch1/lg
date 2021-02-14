package loadgen

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/swtch1/lg/domain"
	"github.com/swtch1/lg/store"
)

func newDispatcher(timeout time.Duration) *dispatcher {
	return &dispatcher{
		pool:    newClientPool(),
		timeout: timeout,
	}
}

// dispatcher is in charge of turning any type into a payload and sending it to the destination.
type dispatcher struct {
	pool    *clientPool
	timeout time.Duration
}

// ErrClientTimeout is returned when our request to the client
// server lasts longer than the defined timeout.
var ErrClientTimeout = fmt.Errorf("request timed out")

func (d *dispatcher) dispatch(r domain.Request, ni nodeInfo, log *logrus.Entry) (domain.Response, metric, error) {
	c := d.pool.get()
	defer d.pool.put(c)

	req, err := http.NewRequest(r.Method, r.Path, bytes.NewReader(r.Body))
	if err != nil {
		return domain.Response{}, metric{}, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	for k, v := range r.Headers {
		for _, h := range v {
			req.Header.Add(k, h)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), d.timeout)
	defer cancel()

	start := time.Now()
	// let's not over-do it testing this thing locally.. add some suspect latency
	randomSleep()
	resp, err := c.Do(req.WithContext(ctx))
	took := time.Since(start)

	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			// use as sentinel to differentiate between client error and internal error like building the request
			return domain.Response{}, metric{}, ErrClientTimeout
		}
		return domain.Response{}, metric{}, fmt.Errorf("failed to send request: %w", err)
	}

	// *must* read full body to avoid TIME_WAIT hanging
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return domain.Response{}, metric{}, fmt.Errorf("failed to read response body: %w", err)
	}
	resp.Body.Close()

	// create response based on call results
	rresp := domain.Response{
		StatusCode: resp.StatusCode,
		Body:       b,
	}

	// store metrics related to the request
	m := metric{
		latency: store.AggLatency{
			NodeID:    ni.ID,
			LatencyMS: uint64(took.Milliseconds()),
		},
	}

	return rresp, m, nil
}

func randomSleep() {
	n := rand.Intn(500)
	time.Sleep(time.Millisecond * time.Duration(n))
}
