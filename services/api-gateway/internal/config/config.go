package config

import "fafnir/api-gateway/internal/clients"

// will use env later

type Config struct {
	PORT    string
	CLIENTS ClientsConfig
}

type ClientsConfig struct {
	AuthClient *clients.AuthClient
}

func NewConfig() *Config {
	return &Config{
		PORT: ":8080",
		CLIENTS: ClientsConfig{
			AuthClient: clients.NewAuthClient("http://auth-service:8081"),
		},
	}
}
