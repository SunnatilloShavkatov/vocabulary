package db

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"vocabulary/backend/libs/shared/config"
)

// NewRedisClient initializes and pings a Redis connection client.
func NewRedisClient(cfg config.RedisConfig) (*redis.Client, error) {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		rdb.Close()
		return nil, fmt.Errorf("ping redis at %s: %w", addr, err)
	}

	return rdb, nil
}
