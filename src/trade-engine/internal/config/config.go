package config

import (
	"fafnir/shared/pkg/redis"
	"fmt"
	"os"
)

type Config struct {
	PORT         string
	NATS         NatsConfig
	StockService StockServiceConfig
	Portfolio    PortfolioServiceConfig
	Cache        redis.CacheConfig
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

func New() *Config {
	return &Config{
		PORT:         fmt.Sprintf(":%s", os.Getenv("SERVICE_PORT")),
		NATS:         newNatsConfig(),
		StockService: newStockServiceConfig(),
		Portfolio:    newPortfolioServiceConfig(),
		Cache:        newRedisConfig(),
	}
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
