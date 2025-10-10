package config

import (
	"flag"
	"fmt"
	"log"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
)

type Config struct {
	Address         string `env:"SERVER_ADDRESS"`
	BaseURL         string `env:"BASE_URL"`
	LogLevel        string `env:"LOG_LEVEL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
}

func New() *Config {
	cfg := &Config{}

	// Парсим флаги командной строки
	flag.StringVar(&cfg.Address, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080", "base url")
	flag.StringVar(&cfg.LogLevel, "l", "info", "log level (debug, info, warn, error)")
	flag.StringVar(&cfg.FileStoragePath, "f", "", "file for saving links")
	flag.StringVar(&cfg.DatabaseDSN, "d", "", "database connection string")
	flag.Parse()

	// Парсим переменные окружения (перезаписываем значения из флагов, если переменные заданы)
	if err := godotenv.Load(); err != nil {
		fmt.Println("Warning: .env file not found, proceeding without it")
	}
	err := env.Parse(cfg)
	if err != nil {
		log.Fatal(err)
	}

	return cfg
}
