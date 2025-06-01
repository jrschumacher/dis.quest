package config

import (
	"os"
)

// Config holds application configuration loaded from environment variables.
type Config struct {
	AppEnv      string
	Port        string
	DatabaseURL string
	PDSEndpoint string
}

// Load loads configuration from environment variables.
func Load() *Config {
	return &Config{
		AppEnv:      os.Getenv("APP_ENV"),
		Port:        os.Getenv("PORT"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		PDSEndpoint: os.Getenv("PDS_ENDPOINT"),
	}
}
