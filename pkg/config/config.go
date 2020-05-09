package config

import (
	"github.com/pelletier/go-toml"
)

// Config contains configuration data required by an application, loaded from a
// configuration file.
type Config struct {
	Address    string `toml:"address" default:":8080"`
	DB         string `toml:"db"`
	Migrations string `toml:"migrations"`

	AWS struct {
		Bucket string `toml:"bucket"`
		Key    string `toml:"key"`
		Secret string `toml:"secret"`
	} `toml:"aws"`
}

// Load loads a given configuration file path into a new Config.
func Load(filepath string) (*Config, error) {
	tree, err := toml.LoadFile(filepath)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := tree.Unmarshal(cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
