package config

import (
	"encoding/json"
	"fmt"
	"os"
)

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

func Load(filepath string) (*Config, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("io read file failed: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("json unmarshal failed: %w", err)
	}

	return &cfg, nil
}
