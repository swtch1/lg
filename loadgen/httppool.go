package loadgen

import (
	"net/http"
	"sync"
)

func newClientPool() *clientPool {
	defaultClient := http.Client{}
	return &clientPool{
		pool: sync.Pool{
			New: func() interface{} { return &defaultClient },
		},
	}
}

// clientPool provides http clients as necessary in a resource efficient way.
type clientPool struct {
	pool sync.Pool
}

func (p *clientPool) get() *http.Client {
	return p.pool.Get().(*http.Client)
}

func (p *clientPool) put(c *http.Client) {
	resetClient(c)
	p.pool.Put(c)
}

// resetClient so it can be reused safely
func resetClient(c *http.Client) {
	c.CloseIdleConnections()
}
