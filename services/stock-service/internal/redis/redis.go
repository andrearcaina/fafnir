package redis

import (
	"context"
	"fafnir/stock-service/internal/config"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	client *redis.Client
	ttl    time.Duration // time to live (expiration duration) for cached items
}

func New(cfg *config.Config) (*Cache, error) {
	var rdb *redis.Client
	var err error

	const maxRetries = 10
	const retryInterval = 5 * time.Second

	for attempt := 1; attempt <= maxRetries; attempt++ {
		rdb = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%s", cfg.Cache.Host, cfg.Cache.Port),
			Password: "", // no password set
			DB:       0,
		})

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		err = rdb.Ping(ctx).Err()
		cancel()

		if err == nil {
			log.Printf("Successfully connected to Redis on attempt %d", attempt)
			return &Cache{
				client: rdb,
				ttl:    5 * time.Minute, // default TTL of 5 minutes
			}, nil
		}

		log.Printf("Attempt %d/%d: Failed to ping Redis: %v", attempt, maxRetries, err)

		rdb.Close()

		if attempt < maxRetries {
			log.Printf("Retrying Redis connection in %v...", retryInterval)
			time.Sleep(retryInterval)
		}
	}

	return nil, fmt.Errorf("failed to connect to Redis after %d attempts: %w", maxRetries, err)
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
