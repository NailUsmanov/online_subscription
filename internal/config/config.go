package config

import (
	"fmt"
	"os"
)

// Config - струтура для получения сервера и DSN БД.
type Config struct {
	ServerAddr  string
	DatabaseDSN string
}

// Load - получает значения из переменных окружения, если они заданы.
func Load() (*Config, error) {
	cfg := &Config{}

	cfg.ServerAddr = os.Getenv("SERVER_ADDR")
	if cfg.ServerAddr == "" {
		cfg.ServerAddr = ":8080"
	}

	cfg.DatabaseDSN = os.Getenv("DATABASE_DSN")
	if cfg.DatabaseDSN == "" {
		return nil, fmt.Errorf("DATABASE_DSN is required")
	}

	return cfg, nil
}
