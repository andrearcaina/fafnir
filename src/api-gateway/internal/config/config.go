package config

import (
	"fafnir/api-gateway/internal/clients"
	"os"
)

// will use env later

type Config struct {
	PORT    string
	CLIENTS ClientsConfig
	PROXY   ProxyConfig
	ENV     EnvConfig
}

type ClientsConfig struct {
	SecurityClient *clients.SecurityClient
	UserClient     *clients.UserClient
	StockClient    *clients.StockClient
}

type ProxyConfig struct {
	TargetURL string
}

type EnvConfig struct {
	JWT string
}

func NewConfig() *Config {
	return &Config{
		PORT: ":8080",
		CLIENTS: ClientsConfig{
			SecurityClient: clients.NewSecurityClient("security-service:8082"),
			UserClient:     clients.NewUserClient("user-service:8083"),
			StockClient:    clients.NewStockClient("stock-service:8084"),
		},
		PROXY: ProxyConfig{
			TargetURL: "http://auth-service:8081/",
		},
		ENV: EnvConfig{
			JWT: os.Getenv("JWT_SECRET_KEY"),
		},
	}
}
