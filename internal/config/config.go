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
// Приоритет источников: json < флаги < переменные окружения.
func New() *Config {
	cfg := &Config{
		TLSCertPath: "cert/cert.pem",
		TLSKeyPath:  "cert/private.pem",
	}

	// Загрузка переменных окружения из .env файла, если он существует
	if err := godotenv.Load(); err != nil {
		fmt.Println("Warning: .env file not found, proceeding without it")
	}

	// Получаем флаги командной строки во временное хранилище, чтобы получить путь до JSON конфига
	fv := parseFlags(cfg)

	// Парсим JSON файл
	configFilePath := findConfigPath(fv.configJSONPath)
	if configFilePath != "" {
		if err := loadFromJSON(cfg, configFilePath); err != nil {
			log.Fatalf("Error loading config from JSON file: %v", err)
		}
	}

	// Применяем флаги командной строки
	applyFlags(cfg, fv)

	// Парсим переменные окружения (высший приоритет)
	if err := loadFromEnv(cfg); err != nil {
		log.Fatalf("Error loading config from environment variables: %v", err)
	}

	return cfg
}

func findConfigPath(flagConfigPath string) string {
	if v, ok := os.LookupEnv("CONFIG"); ok {
		return v
	}
	return flagConfigPath
}

func loadFromJSON(cfg *Config, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewDecoder(file).Decode(cfg)
}

type flagValues struct {
	address         string
	baseURL         string
	logLevel        string
	fileStoragePath string
	databaseDSN     string
	secretKey       string
	auditFile       string
	auditURL        string
	configJSONPath  string
	enableHTTPS     bool
}

func parseFlags(cfg *Config) *flagValues {
	fv := &flagValues{}

	flag.StringVar(&fv.address, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&fv.baseURL, "b", "http://localhost:8080", "base url")
	flag.StringVar(&fv.logLevel, "l", "info", "log level (debug, info, warn, error)")
	flag.StringVar(&fv.fileStoragePath, "f", "", "file for saving links")
	flag.StringVar(&fv.databaseDSN, "d", "", "database connection string")
	flag.StringVar(&fv.secretKey, "k", "", "secret key")
	flag.StringVar(&fv.auditFile, "audit-file", "", "file for audit events")
	flag.StringVar(&fv.auditURL, "audit-url", "", "address for audit events")
	flag.StringVar(&fv.configJSONPath, "c", "", "config file path")
	flag.BoolVar(&fv.enableHTTPS, "s", false, "enable HTTPS")

	flag.Parse()

	return fv
}

func applyFlags(cfg *Config, fv *flagValues) {
	flag.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "a":
			cfg.Address = fv.address
		case "b":
			cfg.BaseURL = fv.baseURL
		case "l":
			cfg.LogLevel = fv.logLevel
		case "f":
			cfg.FileStoragePath = fv.fileStoragePath
		case "d":
			cfg.DatabaseDSN = fv.databaseDSN
		case "k":
			cfg.SecretKey = fv.secretKey
		case "audit-file":
			cfg.AuditFile = fv.auditFile
		case "audit-url":
			cfg.AuditURL = fv.auditURL
		case "s":
			cfg.EnableHTTPS = fv.enableHTTPS
		}
	})
}

func loadFromEnv(cfg *Config) error {
	err := env.Parse(cfg)
	if err != nil {
		log.Fatal(err)
	}

	return nil
}
