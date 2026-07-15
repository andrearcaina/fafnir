package config

import (
	"fmt"
	"os"
	"time"

	"fafnir/shared/pkg/redis"
)

type Config struct {
	PORT         string
	DB           PostgresConfig
	FMP          FMPConfig
	Cache        redis.CacheConfig
	QuoteTTL     time.Duration
	YahooTimeout time.Duration
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
	APIKey  string
	Timeout time.Duration
}

func NewConfig() *Config {
	return &Config{
		PORT:         fmt.Sprintf(":%s", os.Getenv("SERVICE_PORT")),
		DB:           newPostgresConfig(),
		FMP:          newFMPConfig(),
		Cache:        newRedisConfig(),
		QuoteTTL:     durationFromEnv("QUOTE_TTL", time.Minute),
		YahooTimeout: durationFromEnv("YAHOO_TIMEOUT", 10*time.Second),
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
		APIKey:  apiKey,
		Timeout: durationFromEnv("FMP_TIMEOUT", 5*time.Second),
	}
}

func newRedisConfig() redis.CacheConfig {
	host := os.Getenv("REDIS_HOST")
	port := os.Getenv("REDIS_PORT")
	password := os.Getenv("REDIS_PASSWORD")
	db := 0

	return redis.CacheConfig{
		Host:     host,
		Port:     port,
		Password: password,
		DB:       db,
	}
}
