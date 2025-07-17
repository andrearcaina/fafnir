package config

// will use env later

type Config struct {
	PORT   string
	DB_URL string
}

func NewConfig() *Config {
	return &Config{
		PORT:   ":8081",
		DB_URL: "",
	}
}
