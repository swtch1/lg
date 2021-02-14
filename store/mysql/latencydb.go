package mysql

import (
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/swtch1/lg/store"
)

func NewLatencyDB(db *sqlx.DB) *LatencyDB {
	return &LatencyDB{sqlDB: db}
}

type LatencyDB struct {
	sqlDB *sqlx.DB
}

func (db *LatencyDB) GetLatency() ([]store.AggLatency, error) {
	s := squirrel.
		Select(
			"id",
			"node_id",
			"latency_ms",
			"created",
		).
		From(`lg.agg_latency`)

	q, args, err := s.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	ls := make([]store.AggLatency, 0)
	err = db.sqlDB.Select(&ls, q, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get latency records: %w", err)
	}

	return ls, nil
}

func (db *LatencyDB) CreateLatencies(ls []store.AggLatency) error {
	s := squirrel.
		Insert("lg.agg_latency").
		Columns(
			"id",
			"node_id",
			"latency_ms",
			// created applied automatically
		)

	for _, l := range ls {
		s = s.Values(
			l.ID,
			l.NodeID,
			l.LatencyMS,
		)
	}

	q, args, err := s.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = db.sqlDB.Exec(q, args...)
	if err != nil {
		return fmt.Errorf("failed to exec: %w", err)
	}
	return nil
}

// PurgeLatencies from the database.
func (db *LatencyDB) PurgeLatencies() error {
	q, args, err := squirrel.
		Delete("lg.agg_latency").
		Where("1 = 1").
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = db.sqlDB.Exec(q, args...)
	if err != nil {
		return fmt.Errorf("failed to exec: %w", err)
	}
	return nil
}
