# Memoir

Command-line tools and API server for hosting track lists for DJ mixes.

## Accounts

Anyone can register with an email address, a password of 15 to 128 characters,
and a display name. Login uses an HttpOnly, host-only cookie; logout revokes the
session and password reset revokes all sessions. Passwords use Argon2id.

Set `APP_ORIGIN` to the browser application's origin. HTTPS is required outside
localhost. Host the browser app and API on the same site for SameSite cookies.
`SESSION_TTL` defaults to `168h` and accepts up to `720h`.

Fetch `GET /auth/csrf` with credentials, then send `data.csrf_token` in the
`X-CSRF-Token` header on account POST requests. Use `credentials: 'include'`.
The server requires the configured Origin and JSON request bodies.

Account endpoints include registration, login, logout, `/auth/me`, email
verification and resending, forgot password, and reset password. Verification
links expire after 24 hours, and reset links after one hour. Links work once.
The frontend must handle `/verify-email` and `/reset-password` with a token in
the URL fragment and submit that token in the API request body.

Account requests are limited per process and client IP, with additional limits
per email and operation. Forwarded IP headers are not trusted implicitly.
Content writes still use `API_TOKEN` until the ownership cutover layer.

`memoir-user` provisions a verified account using a password from stdin. Its
`-claim-legacy-tracklists` flag assigns currently unowned tracklists to that
account in the same transaction.

## SMTP

Set `SMTP_ADDRESS` to `host:port` and `SMTP_FROM` to the sender address. Set
`SMTP_USERNAME` and `SMTP_PASSWORD` together when authentication is needed.
`SMTP_TLS_MODE` accepts `starttls`, the default, or `tls` for implicit TLS.
STARTTLS must be available when requested, and TLS certificates are verified.
The `none` mode is supported only for a local SMTP server, such as Mailpit on
`localhost:1025`. See `.env.example` for local configuration.

Account emails use a bounded queue in memory, with a five-second delivery
timeout. SMTP delivery happens outside the HTTP request so its latency does not
disclose whether an account exists. Delivery errors are logged without
addresses, credentials, or tokens. Delivery is best effort: a restart, full
queue, or SMTP failure requires the user to request another email. Sessions and
expired account tokens are cleaned up hourly.

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
