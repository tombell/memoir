package config

import (
	"fmt"

	"github.com/pelletier/go-toml"
)

type Config struct {
	Address    string `toml:"address" default:":8080"`
	DB         string `toml:"db"`
	Migrations string `toml:"migrations"`

	API struct {
		Token string `toml:"token"`
	} `toml:"api"`

	AWS struct {
		Bucket string `toml:"bucket"`
		Region string `toml:"region"`
		Key    string `toml:"key"`
		Secret string `toml:"secret"`
	} `toml:"aws"`
}

func Load(filepath string) (*Config, error) {
	tree, err := toml.LoadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("toml load file failed: %w", err)
	}

	var cfg Config
	if err := tree.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("toml unmarshal failed: %w", err)
	}

	return &cfg, nil
}