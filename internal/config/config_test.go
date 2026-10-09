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
