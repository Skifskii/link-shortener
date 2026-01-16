// Package config содержит логику чтения конфигурации из флагов командной строки
// и переменных окружения.
package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
)

// Config содержит параметры конфигурации приложения.
type Config struct {
	Address         string `json:"server_address" env:"SERVER_ADDRESS"`
	BaseURL         string `json:"base_url" env:"BASE_URL"`
	LogLevel        string `json:"log_level" env:"LOG_LEVEL"`
	FileStoragePath string `json:"file_storage_path" env:"FILE_STORAGE_PATH"`
	DatabaseDSN     string `json:"database_dsn" env:"DATABASE_DSN"`
	SecretKey       string `json:"secret_key" env:"SECRET_KEY"`
	AuditFile       string `json:"audit_file" env:"AUDIT_FILE"`
	AuditURL        string `json:"audit_url" env:"AUDIT_URL"`
	EnableHTTPS     bool   `json:"enable_https" env:"ENABLE_HTTPS"`
	TLSCertPath     string
	TLSKeyPath      string
}

// New читает конфигурацию из флагов и переменных окружения и возвращает объект Config.
func New() *Config {
	cfg := &Config{
		TLSCertPath: "cert/cert.pem",
		TLSKeyPath:  "cert/private.pem",
	}

	if err := godotenv.Load(); err != nil {
		fmt.Println("Warning: .env file not found, proceeding without it")
	}

	// 1. JSON файл (низший приоритет)
	configFilePath := findConfigPath()
	if configFilePath != "" {
		if err := loadFromJSON(cfg, configFilePath); err != nil {
			log.Fatalf("Error loading config from JSON file: %v", err)
		}
	}

	// 2. Флаги командной строки
	loadFromFlags(cfg)

	// 3. Переменные окружения (высший приоритет)
	if err := loadFromEnv(cfg); err != nil {
		log.Fatalf("Error loading config from environment variables: %v", err)
	}

	return cfg
}

func findConfigPath() string {
	if v, ok := os.LookupEnv("CONFIG"); ok {
		return v
	}

	var path string
	flag.StringVar(&path, "c", "", "config file path")
	flag.Parse()

	return path
}

func loadFromJSON(cfg *Config, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewDecoder(file).Decode(cfg)
}

func loadFromFlags(cfg *Config) {
	flag.StringVar(&cfg.Address, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080", "base url")
	flag.StringVar(&cfg.LogLevel, "l", "info", "log level (debug, info, warn, error)")
	flag.StringVar(&cfg.FileStoragePath, "f", "", "file for saving links")
	flag.StringVar(&cfg.DatabaseDSN, "d", "", "database connection string")
	flag.StringVar(&cfg.SecretKey, "k", "", "secret key")
	flag.StringVar(&cfg.AuditFile, "audit-file", "", "file for audit events")
	flag.StringVar(&cfg.AuditURL, "audit-url", "", "address for audit events")
	flag.BoolVar(&cfg.EnableHTTPS, "s", false, "enable HTTPS")

	flag.Parse()
}

func loadFromEnv(cfg *Config) error {
	err := env.Parse(cfg)
	if err != nil {
		log.Fatal(err)
	}

	return nil
}
