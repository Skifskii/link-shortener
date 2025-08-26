package config

import "flag"

type Config struct {
	Address string // localhost:8080
	BaseURL string // http://localhost:8080/
}

func New() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.Address, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&cfg.BaseURL, "b", "/", "base url")
	flag.Parse()

	return cfg
}
