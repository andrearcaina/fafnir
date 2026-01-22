package config

import (
	"fafnir/shared/pkg/redis"
	"fmt"
	"os"
)

type Config struct {
	PORT  string
	DB    PostgresConfig
	FMP   FMPConfig
	Cache redis.CacheConfig
}

type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DbName   string
	URL      string
}

type FMPConfig struct {
	APIKey string
}

func NewConfig() *Config {
	return &Config{
		PORT:  fmt.Sprintf(":%s", os.Getenv("SERVICE_PORT")),
		DB:    newPostgresConfig(),
		FMP:   newFMPConfig(),
		Cache: newRedisConfig(),
	}
}

func newPostgresConfig() PostgresConfig {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	dbName := os.Getenv("STOCK_DB")

	return PostgresConfig{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		DbName:   dbName,
		URL: fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
			user,
			password,
			host,
			port,
			dbName,
		),
	}
}

func newFMPConfig() FMPConfig {
	apiKey := os.Getenv("FMP_API_KEY")

	return FMPConfig{
		APIKey: apiKey,
	}
}

func newRedisConfig() redis.CacheConfig {
	host := os.Getenv("REDIS_HOST")
	port := os.Getenv("REDIS_PORT")
	password := os.Getenv("REDIS_PASSWORD")

	return redis.CacheConfig{
		Host:     host,
		Port:     port,
		Password: password,
	}
}
