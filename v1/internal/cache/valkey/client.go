package valkey

import (
	"context"
	"fmt"
	"time"

	"github.com/fehmicorp/fehmi-s3/internal/config"
	"github.com/redis/go-redis/v9"
)

type Client struct {
	client *redis.Client
}

func New(ctx context.Context, cfg config.CacheConfig) (*Client, error) {
	if cfg.Host == "" {
		return nil, fmt.Errorf("valkey host is required")
	}

	if cfg.Port < 1 || cfg.Port > 65535 {
		return nil, fmt.Errorf("invalid valkey port")
	}

	if cfg.Database < 0 {
		return nil, fmt.Errorf("invalid valkey database")
	}

	opts := &redis.Options{
		Addr:            fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:        cfg.Password,
		DB:              cfg.Database,
		PoolSize:        50,
		MinIdleConns:    5,
		DialTimeout:     5 * time.Second,
		ReadTimeout:     3 * time.Second,
		WriteTimeout:    3 * time.Second,
		PoolTimeout:     4 * time.Second,
		ConnMaxIdleTime: 5 * time.Minute,
	}

	client := redis.NewClient(opts)

	cache := &Client{
		client: client,
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := client.Ping(pingCtx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("valkey connection failed: %w", err)
	}

	return cache, nil
}

// Redis returns the underlying Redis-compatible client.
func (c *Client) Redis() *redis.Client {
	return c.client
}

// Ping checks cache connectivity.
func (c *Client) Ping(ctx context.Context) error {
	if c == nil || c.client == nil {
		return fmt.Errorf("valkey client is not initialized")
	}

	return c.client.Ping(ctx).Err()
}

// Get retrieves a cached value.
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	if c == nil || c.client == nil {
		return "", fmt.Errorf("valkey client is not initialized")
	}

	return c.client.Get(ctx, key).Result()
}

// Set stores a cached value with an expiration.
func (c *Client) Set(
	ctx context.Context,
	key string,
	value interface{},
	ttl time.Duration,
) error {
	if c == nil || c.client == nil {
		return fmt.Errorf("valkey client is not initialized")
	}

	return c.client.Set(ctx, key, value, ttl).Err()
}

// Delete removes one or more cache keys.
func (c *Client) Delete(ctx context.Context, keys ...string) error {
	if c == nil || c.client == nil {
		return fmt.Errorf("valkey client is not initialized")
	}

	if len(keys) == 0 {
		return nil
	}

	return c.client.Del(ctx, keys...).Err()
}

// Exists checks whether a cache key exists.
func (c *Client) Exists(ctx context.Context, key string) (bool, error) {
	if c == nil || c.client == nil {
		return false, fmt.Errorf("valkey client is not initialized")
	}

	count, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// Close gracefully closes the cache connection.
func (c *Client) Close() error {
	if c == nil || c.client == nil {
		return nil
	}

	return c.client.Close()
}
