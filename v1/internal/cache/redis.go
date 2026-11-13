package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/fehmicorp/s3object/v1/types"
	"github.com/redis/go-redis/v9"
)

type Cache struct{ Client *redis.Client }

func OpenCache(ctx context.Context, c *types.Cache) (*Cache, error) {
	r := redis.NewClient(&redis.Options{Addr: fmt.Sprintf("%s:%d", c.Host, c.Port), Password: c.Password, DB: c.Database, PoolSize: 20})
	x, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := r.Ping(x).Err(); err != nil {
		_ = r.Close()
		return nil, fmt.Errorf("Valkey ping: %w", err)
	}
	return &Cache{Client: r}, nil
}

func (c *Cache) Close() {
	if c != nil && c.Client != nil {
		_ = c.Client.Close()
	}
}
