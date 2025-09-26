package config

import "fafnir/api-gateway/internal/clients"

// will use env later

type Config struct {
	PORT    string
	CLIENTS ClientsConfig
}

type ClientsConfig struct {
	SecurityClient *clients.SecurityClient
	UserClient     *clients.UserClient
	StockClient    *clients.StockClient
}

func NewConfig() *Config {
	return &Config{
		PORT: ":8080",
		CLIENTS: ClientsConfig{
			SecurityClient: clients.NewSecurityClient("security-service:8082"),
			UserClient:     clients.NewUserClient("user-service:8083"),
			StockClient:    clients.NewStockClient("http://stock-service:8084/stock"),
		},
	}
}
