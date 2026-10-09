package telemetry

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	collector "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	common "go.opentelemetry.io/proto/otlp/common/v1"
	otlplogs "go.opentelemetry.io/proto/otlp/logs/v1"
	"google.golang.org/protobuf/proto"

	"github.com/tombell/memoir/internal/config"
)

func TestNewLoggerDisabled(t *testing.T) {
	var console bytes.Buffer
	base := slog.New(slog.NewTextHandler(&console, nil))
	cfg := &config.Config{}
	cfg.PostHog.Host = "invalid host"

	logger, shutdown, err := NewLogger(context.Background(), cfg, base)
	if err != nil {
		t.Fatal(err)
	}
	if logger != base {
		t.Fatal("disabled export should preserve the existing logger")
	}
	logger.Info("console only")
	if !strings.Contains(console.String(), "console only") {
		t.Fatal("log did not reach the console")
	}
	if err := shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestNewLoggerInvalidHost(t *testing.T) {
	for _, host := range []string{
		"", "us.i.posthog.com", "://bad", "ftp://posthog.example",
		"https://user:secret@posthog.example", "https://posthog.example?token=secret",
		"https://posthog.example#logs",
	} {
		t.Run(host, func(t *testing.T) {
			cfg := &config.Config{}
			cfg.PostHog.APIKey = "phc_test"
			cfg.PostHog.Host = host
			base := slog.New(slog.NewTextHandler(io.Discard, nil))

			_, _, err := NewLogger(context.Background(), cfg, base)
			if err == nil {
				t.Fatal("expected invalid host to be rejected")
			}
			if strings.Contains(err.Error(), "secret") {
				t.Fatal("configuration error leaked credentials")
			}
		})
	}
}

func TestNewLoggerExportsAndFlushes(t *testing.T) {
	for _, tt := range []struct {
		name        string
		serviceName string
		attributes  string
		prefix      string
		wantService string
	}{
		{name: "defaults", wantService: "memoir"},
		{
			name:        "configured",
			serviceName: "memoir-api",
			attributes:  "deployment.environment.name=production",
			prefix:      "/gateway/",
			wantService: "memoir-api",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("OTEL_SERVICE_NAME", tt.serviceName)
			t.Setenv("OTEL_RESOURCE_ATTRIBUTES", tt.attributes)

			requests := make(chan *collector.ExportLogsServiceRequest, 16)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost || r.URL.Path != strings.TrimRight(tt.prefix, "/")+"/i/v1/logs" {
					t.Errorf("unexpected export request: %s %s", r.Method, r.URL.Path)
				}
				if r.Header.Get("Authorization") != "Bearer phc_test" {
					t.Error("missing project token bearer authentication")
				}
				if r.Header.Get("Content-Type") != "application/x-protobuf" {
					t.Error("export is not OTLP protobuf")
				}
				body, err := io.ReadAll(r.Body)
				if err != nil {
					t.Error(err)
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				request := &collector.ExportLogsServiceRequest{}
				if err := proto.Unmarshal(body, request); err != nil {
					t.Error(err)
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				requests <- request
				w.Header().Set("Content-Type", "application/x-protobuf")
			}))
			defer server.Close()

			var console bytes.Buffer
			base := slog.New(slog.NewJSONHandler(&console, nil))
			cfg := &config.Config{}
			cfg.PostHog.APIKey = "phc_test"
			cfg.PostHog.Host = server.URL + tt.prefix

			logger, shutdown, err := NewLogger(context.Background(), cfg, base)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				_ = shutdown(ctx)
			})

			logger.Info("starting api server")
			logger.With("rid", "req-123").WithGroup("http").With("method", "GET").Error(
				"request failed", "status", 500, "err", errors.New("storage unavailable"),
			)

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := shutdown(ctx); err != nil {
				t.Fatalf("flush logs on shutdown: %v", err)
			}
			close(requests)

			var records []*otlplogs.LogRecord
			for request := range requests {
				for _, resourceLogs := range request.ResourceLogs {
					attrs := resourceLogs.Resource.Attributes
					if got := findAttribute(attrs, "service.name").GetStringValue(); got != tt.wantService {
						t.Errorf("service.name = %q, want %q", got, tt.wantService)
					}
					if tt.attributes != "" && findAttribute(attrs, "deployment.environment.name").GetStringValue() != "production" {
						t.Error("missing configured resource attributes")
					}
					for _, scopeLogs := range resourceLogs.ScopeLogs {
						records = append(records, scopeLogs.LogRecords...)
					}
				}
			}
			if len(records) != 2 {
				t.Fatalf("exported %d logs, want 2", len(records))
			}
			if records[0].Body.GetStringValue() != "starting api server" || records[0].SeverityText != "INFO" {
				t.Error("startup log message or severity was lost")
			}
			failure := records[1]
			if failure.Body.GetStringValue() != "request failed" || failure.SeverityText != "ERROR" {
				t.Error("error log message or severity was lost")
			}
			if findAttribute(failure.Attributes, "rid").GetStringValue() != "req-123" {
				t.Error("request ID was lost")
			}
			httpAttrs := findAttribute(failure.Attributes, "http").GetKvlistValue().GetValues()
			if findAttribute(httpAttrs, "method").GetStringValue() != "GET" ||
				findAttribute(httpAttrs, "status").GetIntValue() != 500 {
				t.Error("grouped request attributes were lost")
			}
			if findAttribute(failure.Attributes, "exception.message").GetStringValue() != "storage unavailable" ||
				findAttribute(failure.Attributes, "exception.type").GetStringValue() == "" {
				t.Error("error message or type was lost")
			}
			if !strings.HasSuffix(findAttribute(failure.Attributes, "code.file.path").GetStringValue(), "logs_test.go") {
				t.Error("log source location was lost")
			}
			if !strings.Contains(console.String(), "starting api server") ||
				!strings.Contains(console.String(), "request failed") ||
				!strings.Contains(console.String(), "req-123") {
				t.Error("logs or request attributes did not reach the console")
			}
		})
	}
}

func findAttribute(attrs []*common.KeyValue, key string) *common.AnyValue {
	for _, attr := range attrs {
		if attr.Key == key {
			return attr.Value
		}
	}
	return nil
}
