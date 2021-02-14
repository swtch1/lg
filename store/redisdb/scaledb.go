package redisdb

import (
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v7"
)

func NewScaleDB(rc *redis.Client, scaleKey string) *ScaleDB {
	return &ScaleDB{
		rc:  rc,
		key: scaleKey,
	}
}

type ScaleDB struct {
	rc  *redis.Client
	key string
}

func (db *ScaleDB) SetScaleFactor(f float64) error {
	_, err := db.rc.Set(db.key, f, time.Hour).Result()
	if err != nil {
		return fmt.Errorf("failed to set: %w", err)
	}
	return nil
}

func (db *ScaleDB) GetScaleFactor() (float64, error) {
	r, err := db.rc.Get(db.key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get key %s: %w", db.key, err)
	}
	f, err := strconv.ParseFloat(r, 64)
	if err != nil {
		return 0, fmt.Errorf("scale factor is not a float: %w", err)
	}
	return f, nil
}
