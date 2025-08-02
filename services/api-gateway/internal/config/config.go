package config

import "fafnir/api-gateway/internal/clients"

// will use env later

type Config struct {
	PORT       string
	AuthClient *clients.AuthClient
}

func NewConfig() *Config {
	return &Config{
		PORT:       ":8080",
		AuthClient: clients.NewAuthClient("http://fafnir-auth-service-1:8081/"),
	}
}
