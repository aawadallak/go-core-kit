package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

func newDefaultOptions() *redis.Options {
	return &redis.Options{
		Addr:            "localhost:6379",
		Password:        "",
		DB:              0,
		MaxRetries:      3,
		MinRetryBackoff: time.Millisecond * 8,
		MaxRetryBackoff: time.Second * 512,
		DialTimeout:     time.Second * 5,
		ReadTimeout:     time.Second * 3,
		WriteTimeout:    time.Second * 3,
		PoolSize:        10,
		MinIdleConns:    5,
		PoolTimeout:     time.Second * 4,
		MaxIdleConns:    10,
	}
}

// newClient creates a new Redis client with the given configuration and verifies the connection.
func newClient(ctx context.Context, config *redis.Options) (*redis.Client, error) {
	client := redis.NewClient(config)

	// Attempt to ping the Redis server with a small delay
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return client, nil
}
