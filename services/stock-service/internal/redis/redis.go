package redis

import (
	"context"
	"fafnir/stock-service/internal/config"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	client *redis.Client
	ttl    time.Duration // time to live (expiration duration) for cached items
}

func New(config *config.Config) (*Cache, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", config.Cache.Host, config.Cache.Port),
		Password: "", // no password set
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping failed: %v", err)
	}

	return &Cache{
		client: rdb,
		ttl:    5 * time.Minute, // default TTL of 5 minutes
	}, nil
}

func (c *Cache) Close() error {
	return c.client.Close()
}

// Redis uses 3 main commands (since it is a key-val store):

// Set - to set a value for a key with an expiration time
func (c *Cache) Set(ctx context.Context, key string, value string) error {
	return c.client.Set(ctx, key, value, c.ttl).Err()
}

// Get - to get the value for a key
func (c *Cache) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

// Del - to delete a key
func (c *Cache) Del(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}
