package config

import (
	"fmt"
	"os"
)

type Config struct {
	PORT         string
	NATS         NatsConfig
	StockService StockServiceConfig
	Portfolio    PortfolioServiceConfig
}

type NatsConfig struct {
	Host string
	Port string
	URL  string
}

type StockServiceConfig struct {
	URL string
}

func New() *Config {
	return &Config{
		PORT:         fmt.Sprintf(":%s", os.Getenv("SERVICE_PORT")),
		NATS:         newNatsConfig(),
		StockService: newStockServiceConfig(),
		Portfolio:    newPortfolioServiceConfig(),
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

func newStockServiceConfig() StockServiceConfig {
	host := os.Getenv("STOCK_SERVICE_HOST")
	port := os.Getenv("STOCK_SERVICE_PORT")

	return StockServiceConfig{
		URL: fmt.Sprintf("%s:%s", host, port),
	}
}

type PortfolioServiceConfig struct {
	URL string
}

func newPortfolioServiceConfig() PortfolioServiceConfig {
	host := os.Getenv("PORTFOLIO_SERVICE_HOST")
	port := os.Getenv("PORTFOLIO_SERVICE_PORT")

	return PortfolioServiceConfig{
		URL: fmt.Sprintf("%s:%s", host, port),
	}
}
