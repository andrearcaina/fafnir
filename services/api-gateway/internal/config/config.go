package config

// will use env later

type Config struct {
	PORT    string
	CLIENTS ClientsConfig
}

type ClientsConfig struct {
	// UserClient *clients.UserClient
}

func NewConfig() *Config {
	return &Config{
		PORT:    ":8080",
		CLIENTS: ClientsConfig{
			// UserClient: clients.NewUserClient("http://user-service:8082"),
		},
	}
}
