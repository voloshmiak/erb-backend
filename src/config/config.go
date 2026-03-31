package config

import (
	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
)

type Config struct {
	Port string `env:"PORT" envDefault:"8080"`
}

func New() (*Config, error) {
	err := godotenv.Load(".env")
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
