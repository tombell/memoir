package config

import (
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config contains the configuration values for various parts of the
// application.
type Config struct {
	Address string
	DB      string
	Auth    AuthConfig
	SMTP    SMTPConfig

	AWS struct {
		Bucket string
		Region string
		Key    string
		Secret string
	}

	PostHog struct {
		APIKey string
		Host   string
	}
}

type AuthConfig struct {
	Origin       string
	CookieSecure bool
	SessionTTL   time.Duration
}

func (c AuthConfig) SessionCookieName() string {
	if c.CookieSecure {
		return "__Host-memoir_session"
	}
	return "memoir_session"
}

func (c AuthConfig) CSRFCookieName() string {
	if c.CookieSecure {
		return "__Host-memoir_csrf"
	}
	return "memoir_csrf"
}

type SMTPConfig struct {
	Address  string
	Username string
	Password string
	From     string
	TLSMode  string
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
		PostHog: struct {
			APIKey string
			Host   string
		}{
			APIKey: os.Getenv("POSTHOG_API_KEY"),
			Host:   getEnv("POSTHOG_HOST", "https://us.i.posthog.com"),
		},
	}
	if err := cfg.loadAccounts(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) loadAccounts() error {
	origin, err := url.Parse(getEnv("APP_ORIGIN", "http://localhost:3000"))
	if err != nil || origin.Hostname() == "" || origin.User != nil || origin.RawQuery != "" || origin.Fragment != "" || (origin.Path != "" && origin.Path != "/") || (origin.Scheme != "http" && origin.Scheme != "https") {
		return fmt.Errorf("APP_ORIGIN must be an HTTP or HTTPS origin")
	}
	if origin.Scheme == "http" && !isLoopback(origin.Hostname()) {
		return fmt.Errorf("APP_ORIGIN requires HTTPS except on localhost")
	}
	ttl, err := time.ParseDuration(getEnv("SESSION_TTL", "168h"))
	if err != nil || ttl <= 0 || ttl > 30*24*time.Hour {
		return fmt.Errorf("SESSION_TTL must be positive and at most 720h")
	}
	c.Auth = AuthConfig{Origin: origin.Scheme + "://" + origin.Host, CookieSecure: origin.Scheme == "https", SessionTTL: ttl}
	c.SMTP = SMTPConfig{
		Address:  getEnv("SMTP_ADDRESS", "localhost:1025"),
		Username: os.Getenv("SMTP_USERNAME"), Password: os.Getenv("SMTP_PASSWORD"),
		From: getEnv("SMTP_FROM", "memoir@localhost"), TLSMode: getEnv("SMTP_TLS_MODE", "starttls"),
	}
	host, port, err := net.SplitHostPort(c.SMTP.Address)
	portNumber, portErr := strconv.Atoi(port)
	if err != nil || host == "" || portErr != nil || portNumber < 1 || portNumber > 65535 {
		return fmt.Errorf("SMTP_ADDRESS must contain a host and port")
	}
	if c.SMTP.TLSMode != "starttls" && c.SMTP.TLSMode != "tls" && c.SMTP.TLSMode != "none" {
		return fmt.Errorf("SMTP_TLS_MODE must be starttls, tls, or none")
	}
	if c.SMTP.TLSMode == "none" && !isLoopback(host) {
		return fmt.Errorf("unencrypted SMTP is only supported on localhost")
	}
	from, err := mail.ParseAddress(c.SMTP.From)
	if err != nil || strings.ContainsAny(c.SMTP.From, "\r\n") {
		return fmt.Errorf("SMTP_FROM must be an email address")
	}
	c.SMTP.From = from.Address
	if (c.SMTP.Username == "") != (c.SMTP.Password == "") {
		return fmt.Errorf("SMTP_USERNAME and SMTP_PASSWORD must be set together")
	}
	return nil
}

func isLoopback(host string) bool {
	ip := net.ParseIP(host)
	return host == "localhost" || (ip != nil && ip.IsLoopback())
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
