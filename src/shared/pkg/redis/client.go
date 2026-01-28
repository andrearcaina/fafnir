package redis

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	client *redis.Client
	ttl    time.Duration // time to live (expiration duration) for cached items
}

type CacheConfig struct {
	Host     string
	Port     string
	Password string
}

// New: initializes a new Redis client with retry logic
func New(config CacheConfig) (*Cache, error) {
	var rdb *redis.Client
	var err error

	const maxRetries = 10
	const retryInterval = 5 * time.Second

	for attempt := 1; attempt <= maxRetries; attempt++ {
		rdb = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%s", config.Host, config.Port),
			Password: config.Password,
			DB:       0,
		})

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		err = rdb.Ping(ctx).Err()
		cancel()

		if err == nil {
			log.Println("Successfully connected to Redis", "attempt", attempt, "addr", rdb.Options().Addr)
			return &Cache{
				client: rdb,
				ttl:    5 * time.Minute,
			}, nil
		}

		log.Println("Failed to ping Redis", "attempt", attempt, "error", err, "retry_in", retryInterval.String())

		err = rdb.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to close Redis connection: %w", err)
		}

		if attempt < maxRetries {
			time.Sleep(retryInterval)
		}
	}

	return nil, fmt.Errorf("failed to connect to Redis after %d attempts: %w", maxRetries, err)
}

// Set: to set a value for a key with an expiration time
func (c *Cache) Set(ctx context.Context, key string, value string) error {
	return c.client.Set(ctx, key, value, c.ttl).Err()
}

// Get: to get the value for a key
func (c *Cache) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

// Del: to delete a key
func (c *Cache) Del(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

// MGet: to get multiple values for multiple keys
func (c *Cache) MGet(ctx context.Context, keys []string) ([]interface{}, error) {
	if len(keys) == 0 {
		return nil, errors.New("no keys provided")
	}
	return c.client.MGet(ctx, keys...).Result()
}

func (c *Cache) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}
