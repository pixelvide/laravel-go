package config

import (
	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

// Load reads the .env file and populates the Config struct
func Load() (*Config, error) {
	// Attempt to load .env file, but don't fail if missing (environment might be set otherwise)
	_ = godotenv.Load()

	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
