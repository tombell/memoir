package telemetry

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"

	"github.com/tombell/memoir/internal/config"
)

// NewLogger adds batched PostHog log export to the existing logger when a
// project token is configured. The caller must shut it down before exiting.
func NewLogger(ctx context.Context, cfg *config.Config, logger *slog.Logger) (*slog.Logger, func(context.Context) error, error) {
	shutdown := func(context.Context) error { return nil }
	if cfg.PostHog.APIKey == "" {
		return logger, shutdown, nil
	}

	host, err := url.Parse(cfg.PostHog.Host)
	if err != nil || host.Host == "" || (host.Scheme != "http" && host.Scheme != "https") ||
		host.User != nil || host.RawQuery != "" || host.Fragment != "" {
		return nil, nil, fmt.Errorf("POSTHOG_HOST must be an HTTP or HTTPS URL without credentials, query parameters, or a fragment")
	}
	host.Path = strings.TrimRight(host.Path, "/") + "/i/v1/logs"
	host.RawPath = ""

	res, err := resource.New(ctx,
		resource.WithAttributes(attribute.String("service.name", "memoir")),
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("create log resource: %w", err)
	}

	exporter, err := otlploghttp.New(ctx,
		otlploghttp.WithEndpointURL(host.String()),
		otlploghttp.WithHeaders(map[string]string{
			"Authorization": "Bearer " + cfg.PostHog.APIKey,
		}),
		otlploghttp.WithTimeout(5*time.Second),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("create PostHog log exporter: %w", err)
	}

	provider := sdklog.NewLoggerProvider(
		sdklog.WithResource(res),
		sdklog.WithProcessor(sdklog.NewBatchProcessor(exporter, sdklog.WithExportTimeout(5*time.Second))),
	)
	handler := otelslog.NewHandler("github.com/tombell/memoir",
		otelslog.WithLoggerProvider(provider),
		otelslog.WithSource(true),
	)

	return slog.New(slog.NewMultiHandler(logger.Handler(), handler)), provider.Shutdown, nil
}
