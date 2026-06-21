package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port      string
	DBUrl     string
	JWTSecret string
	Env       string
}

func Load() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: No .env file found, relying on system environment variables")
	}

	return &Config{
		Port:      os.Getenv("PORT"),
		DBUrl:     os.Getenv("DB_URL"),
		JWTSecret: os.Getenv("JWTSecret"),
		Env:       os.Getenv("ENV"),
	}
}
