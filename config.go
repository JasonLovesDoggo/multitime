package main

import (
	"fmt"
	"os"

	"github.com/pelletier/go-toml/v2"
)

type Backend struct {
	Name      string `toml:"name"`
	URL       string `toml:"url"`
	APIKey    string `toml:"api_key"`
	IsPrimary bool   `toml:"is_primary"`
}

type Config struct {
	Port     int       `toml:"port"`
	Debug    bool      `toml:"debug"`
	Backends []Backend `toml:"backends"`
}

var config *Config

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if cfg.Port == 0 {
		cfg.Port = 3000
	}

	primaryCount := 0
	for _, b := range cfg.Backends {
		if b.IsPrimary {
			primaryCount++
		}
	}
	if primaryCount != 1 {
		return nil, fmt.Errorf("exactly one backend must be marked as primary")
	}

	return &cfg, nil
}
