package config

import (
	"encoding/json"
	"fmt"
	"net"
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

// Load reads the configuration file at the given path and returns a Config instance.
// If the HOST or PORT environment variables are set, they override the host and port in the address.
// If DATABASE_URL is set, it overrides the database connection URL.
func Load(filepath string) (*Config, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("could not read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("could not parse config file: %w", err)
	}

	host, port, err := net.SplitHostPort(cfg.Address)
	if err != nil {
		host = ""
		port = cfg.Address

		if len(port) > 0 && port[0] == ':' {
			port = port[1:]
		}
	}

	if envHost := os.Getenv("HOST"); envHost != "" {
		host = envHost
	}
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}

	cfg.Address = net.JoinHostPort(host, port)

	if envDB := os.Getenv("DATABASE_URL"); envDB != "" {
		cfg.DB = envDB
	}

	return &cfg, nil
}
