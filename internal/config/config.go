package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config contains the configuration values for various parts of the
// application.
type Config struct {
	Address string `json:"address"`
	DB      string `json:"db"`

	API struct {
		Token string `json:"token"`
	} `json:"api"`

	AWS struct {
		Bucket string `json:"bucket"`
		Region string `json:"region"`
		Key    string `json:"key"`
		Secret string `json:"secret"`
	} `json:"aws"`
}

// Load reads the configuration file at the given path and returns a Config
// instance.
func Load(filepath string) (*Config, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("could not read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("could not parse config file: %w", err)
	}

	return &cfg, nil
}
