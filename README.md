# Memoir

Command-line tools and API server for hosting track lists for DJ mixes.

## Logs

The API writes logs to stderr. Set `POSTHOG_API_KEY` to your PostHog project
token to also export these logs through OpenTelemetry. Log export is disabled
when the token is unset.

Set `POSTHOG_HOST` to `https://us.i.posthog.com` for US Cloud or
`https://eu.i.posthog.com` for EU Cloud. The default is US Cloud. Use a project
token starting with `phc_`, as described in the
[PostHog Go Logs setup](https://posthog.com/docs/logs/installation/go).

`OTEL_SERVICE_NAME` sets the service name shown in PostHog and defaults to
`memoir`. Use `OTEL_RESOURCE_ATTRIBUTES` for other metadata, for example
`deployment.environment.name=production`.

Requests and application errors use the existing `slog` logger. Logs are sent
in batches to `/i/v1/logs` and flushed during shutdown, with a five-second
timeout. To verify delivery, start the API with the token configured, make a
request, and check PostHog Logs for the `memoir` service.
