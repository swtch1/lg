package store

// AggLatency contains aggregated latency measurements.
type AggLatency struct {
	ID        int    `db:"id"`
	NodeID    string `db:"node_id"`
	LatencyMS uint64 `db:"latency_ms"`
	Created   string `db:"created"`
}
