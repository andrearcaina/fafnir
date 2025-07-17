package config

// will use env later

type Config struct {
	PORT   string
	DB_URL string
}

func NewConfig() *Config {
	return &Config{
		PORT:   ":8082",
		DB_URL: "",
	}
}
