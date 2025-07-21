package config

// will use env later

type Config struct {
	PORT string
}

func NewConfig() *Config {
	return &Config{
		PORT: ":8080",
	}
}
