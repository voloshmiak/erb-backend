package config

import (
	"log"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
)

type Config struct {
	Port     string `env:"PORT" envDefault:"8080"`
	Database struct {
		Name         string `env:"DB_NAME,required"`
		Password     string `env:"DB_PASSWORD"`
		User         string `env:"DB_USER,required"`
		UserPassword string `env:"DB_USER_PASSWORD,required"`
		Host         string `env:"DB_HOST,required"`
	}
}

func New() (*Config, error) {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("No .env file found, relying on environment variables")
	}

	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
