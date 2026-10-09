package config

import "testing"

func TestLoadPostHog(t *testing.T) {
	for _, key := range []string{"DATABASE_URL", "API_TOKEN", "AWS_BUCKET", "AWS_REGION", "AWS_KEY", "AWS_SECRET"} {
		t.Setenv(key, "test")
	}

	for _, tt := range []struct {
		name     string
		apiKey   string
		host     string
		wantHost string
	}{
		{name: "disabled", wantHost: "https://us.i.posthog.com"},
		{name: "EU Cloud", apiKey: "phc_test", host: "https://eu.i.posthog.com", wantHost: "https://eu.i.posthog.com"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("POSTHOG_API_KEY", tt.apiKey)
			t.Setenv("POSTHOG_HOST", tt.host)
			cfg, err := Load()
			if err != nil {
				t.Fatal(err)
			}
			if cfg.PostHog.APIKey != tt.apiKey || cfg.PostHog.Host != tt.wantHost {
				t.Errorf("unexpected PostHog config: %+v", cfg.PostHog)
			}
		})
	}
}

func TestAccountConfiguration(t *testing.T) {
	for _, key := range []string{"APP_ORIGIN", "SESSION_TTL", "SMTP_ADDRESS", "SMTP_USERNAME", "SMTP_PASSWORD", "SMTP_FROM", "SMTP_TLS_MODE"} {
		t.Setenv(key, "")
	}
	for _, tt := range []struct {
		key, value string
		valid      bool
	}{
		{"APP_ORIGIN", "https://app.example.test", true},
		{"APP_ORIGIN", "http://app.example.test", false},
		{"APP_ORIGIN", "https://app.example.test/path", false},
		{"APP_ORIGIN", "https://user:secret@app.example.test", false},
		{"SESSION_TTL", "0s", false},
		{"SESSION_TTL", "721h", false},
		{"SMTP_ADDRESS", "missing-port", false},
		{"SMTP_TLS_MODE", "none", true},
		{"SMTP_TLS_MODE", "invalid", false},
		{"SMTP_FROM", "sender@example.test\r\nBcc: other@example.test", false},
		{"SMTP_USERNAME", "user", false},
	} {
		t.Run(tt.key+"="+tt.value, func(t *testing.T) {
			t.Setenv(tt.key, tt.value)
			cfg := &Config{}
			err := cfg.loadAccounts()
			if (err == nil) != tt.valid {
				t.Fatalf("valid=%v, error=%v", tt.valid, err)
			}
		})
	}
}
