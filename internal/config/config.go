package config

import (
	"fmt"
	"net"
	"os"

	"github.com/joho/godotenv"
)

// Config contains the configuration values for various parts of the
// application.
type Config struct {
	Address string
	DB      string

	API struct {
		Token string
	}

	AWS struct {
		Bucket string
		Region string
		Key    string
		Secret string
	}
}

// Load reads environment variables from a .env file if present, then
// populates and returns a Config with values from the environment.
// Returns an error if required environment variables are not set.
func Load() (*Config, error) {
	_ = godotenv.Load()

	host := getEnv("HOST", "127.0.0.1")
	port := getEnv("PORT", "8080")

	db, err := requireEnv("DATABASE_URL")
	if err != nil {
		return nil, err
	}

	apiToken, err := requireEnv("API_TOKEN")
	if err != nil {
		return nil, err
	}

	awsBucket, err := requireEnv("AWS_BUCKET")
	if err != nil {
		return nil, err
	}

	awsRegion, err := requireEnv("AWS_REGION")
	if err != nil {
		return nil, err
	}

	awsKey, err := requireEnv("AWS_KEY")
	if err != nil {
		return nil, err
	}

	awsSecret, err := requireEnv("AWS_SECRET")
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		Address: net.JoinHostPort(host, port),
		DB:      db,
		API: struct {
			Token string
		}{
			Token: apiToken,
		},
		AWS: struct {
			Bucket string
			Region string
			Key    string
			Secret string
		}{
			Bucket: awsBucket,
			Region: awsRegion,
			Key:    awsKey,
			Secret: awsSecret,
		},
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func requireEnv(key string) (string, error) {
	if value := os.Getenv(key); value != "" {
		return value, nil
	}
	return "", fmt.Errorf("required environment variable %s is not set", key)
}
