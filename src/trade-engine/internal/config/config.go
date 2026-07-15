package config

import (
	"fmt"
	"os"
	"time"

	"fafnir/shared/pkg/redis"
)

type Config struct {
	PORT         string
	NATS         NatsConfig
	StockService StockServiceConfig
	Portfolio    PortfolioServiceConfig
	Cache        redis.CacheConfig
	FX           FXConfig
}

type NatsConfig struct {
	Host string
	Port string
	URL  string
}

type StockServiceConfig struct {
	URL string
}

type PortfolioServiceConfig struct {
	URL string
}

type FXConfig struct {
	BaseURL string
	Timeout time.Duration
	TTL     time.Duration
}

func New() *Config {
	return &Config{
		PORT:         fmt.Sprintf(":%s", os.Getenv("SERVICE_PORT")),
		NATS:         newNatsConfig(),
		StockService: newStockServiceConfig(),
		Portfolio:    newPortfolioServiceConfig(),
		Cache:        newRedisConfig(),
		FX:           newFXConfig(),
	}
}

func newFXConfig() FXConfig {
	baseURL := os.Getenv("FX_API_URL")
	if baseURL == "" {
		baseURL = "https://api.frankfurter.dev"
	}

	return FXConfig{
		BaseURL: baseURL,
		Timeout: durationFromEnv("FX_TIMEOUT", 5*time.Second),
		TTL:     durationFromEnv("FX_CACHE_TTL", 12*time.Hour),
	}
}

func durationFromEnv(name string, fallback time.Duration) time.Duration {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}

	duration, err := time.ParseDuration(value)
	if err != nil || duration <= 0 {
		return fallback
	}

	return duration
}

func newNatsConfig() NatsConfig {
	host := os.Getenv("NATS_HOST")
	port := os.Getenv("NATS_PORT")

	return NatsConfig{
		Host: host,
		Port: port,
		URL:  fmt.Sprintf("nats://%s:%s", host, port),
	}
}

func newRedisConfig() redis.CacheConfig {
	host := os.Getenv("REDIS_HOST")
	port := os.Getenv("REDIS_PORT")
	password := os.Getenv("REDIS_PASSWORD")
	db := 1

	return redis.CacheConfig{
		Host:     host,
		Port:     port,
		Password: password,
		DB:       db,
	}
}

func newStockServiceConfig() StockServiceConfig {
	host := os.Getenv("STOCK_SERVICE_HOST")
	port := os.Getenv("STOCK_SERVICE_PORT")

	return StockServiceConfig{
		URL: fmt.Sprintf("%s:%s", host, port),
	}
}

func newPortfolioServiceConfig() PortfolioServiceConfig {
	host := os.Getenv("PORTFOLIO_SERVICE_HOST")
	port := os.Getenv("PORTFOLIO_SERVICE_PORT")

	return PortfolioServiceConfig{
		URL: fmt.Sprintf("%s:%s", host, port),
	}
}
