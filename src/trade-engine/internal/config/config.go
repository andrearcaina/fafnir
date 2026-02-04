package config

import (
	"fmt"
	"os"
)

type Config struct {
	NAME         string
	PORT         string
	NATS         NatsConfig
	StockService StockServiceConfig
}

type NatsConfig struct {
	URL string
}

type StockServiceConfig struct {
	URL string
}

func New() *Config {
	return &Config{
		NAME:         os.Getenv("SERVICE_NAME"),
		PORT:         fmt.Sprintf(":%s", os.Getenv("SERVICE_PORT")),
		NATS:         newNatsConfig(),
		StockService: newStockServiceConfig(),
	}
}

func newNatsConfig() NatsConfig {
	host := os.Getenv("NATS_HOST")
	port := os.Getenv("NATS_PORT")

	return NatsConfig{
		URL: fmt.Sprintf("nats://%s:%s", host, port),
	}
}

func newStockServiceConfig() StockServiceConfig {
	host := os.Getenv("STOCK_SERVICE_HOST")
	port := os.Getenv("STOCK_SERVICE_PORT")

	return StockServiceConfig{
		URL: fmt.Sprintf("%s:%s", host, port),
	}
}
