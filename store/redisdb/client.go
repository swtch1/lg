package redisdb

import (
	"fmt"

	"github.com/go-redis/redis/v7"
)

func NewClient(cfg Config, db int) (*redis.Client, error) {
	const maxDB = 15
	if db < 0 || db > maxDB {
		return nil, fmt.Errorf("db must be between 0 and %d", maxDB)
	}

	address := cfg.Host + ":" + cfg.Port
	options := &redis.Options{
		Addr:       address,
		Password:   cfg.Password,
		DB:         db,
		MaxRetries: 3,
	}
	rc := redis.NewClient(options)
	if _, err := rc.Ping().Result(); err != nil {
		return nil, fmt.Errorf("failed to ping redis server: %w", err)
	}
	return rc, nil
}
