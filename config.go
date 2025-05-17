package main

import (
	"fmt"
	"github.com/pelletier/go-toml/v2"
	"os"
)

type Backend struct {
	Name      string `toml:"name"`
	URL       string `toml:"url"`
	APIKey    string `toml:"api_key"`
	IsPrimary bool   `toml:"is_primary,omitempty"` // omitempty allows the field to be optional
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
	if primaryCount == 0 {
		return nil, fmt.Errorf("no primary backend set - exactly one backend must be marked with is_primary = true")
	}
	if primaryCount > 1 {
		return nil, fmt.Errorf("multiple primary backends found - exactly one backend must be marked with is_primary = true")
	}

	return &cfg, nil
}
