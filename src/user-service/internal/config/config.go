package config

import (
	"fmt"
	"os"
)

type Config struct {
	PORT string
	DB   PostgresConfig
	NATS NatsConfig
}

type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DbName   string
	URL      string
}

type NatsConfig struct {
	Host string
	Port string
	URL  string
}

func NewConfig() *Config {
	return &Config{
		PORT: fmt.Sprintf(":%s", os.Getenv("SERVICE_PORT")),
		DB:   newPostgresConfig(),
		NATS: newNatsConfig(),
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

func newPostgresConfig() PostgresConfig {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	dbName := os.Getenv("USER_DB")

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
