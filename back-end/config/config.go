package config

import (
	"errors"
	"os"
)

type Config struct {
	Env      string
	Server   ServerConfig
	Database DatabaseConfig
}

type ServerConfig struct {
	Port string
}

type DatabaseConfig struct {
	DSN string
}

// Load lê as variáveis de ambiente (já carregadas do .env pelo main).
func Load() (*Config, error) {
	cfg := &Config{
		Env: getEnv("APP_ENV", "development"),
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
		},
		Database: DatabaseConfig{
			DSN: os.Getenv("CONNECTION_STRING"),
		},
	}

	if cfg.Database.DSN == "" {
		return nil, errors.New("config: CONNECTION_STRING não definida")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
