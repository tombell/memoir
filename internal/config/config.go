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

// Load reads the configuration file at the given path and returns a Config
// instance. If the HOST or PORT environment variables are set, they override
// the host and port in the address.
func Load(filepath string) (*Config, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("could not read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("could not parse config file: %w", err)
	}

	// Parse current address to get host and port
	host, port, err := net.SplitHostPort(cfg.Address)
	if err != nil {
		// If parsing fails, assume it's just a port (e.g., ":8080")
		host = ""
		port = cfg.Address
		if len(port) > 0 && port[0] == ':' {
			port = port[1:]
		}
	}

	// Override with environment variables if set
	if envHost := os.Getenv("HOST"); envHost != "" {
		host = envHost
	}
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}

	// Reconstruct address
	cfg.Address = net.JoinHostPort(host, port)

	return &cfg, nil
}
