# Memoir

Command-line tools and API server for hosting track lists for DJ mixes.

## Accounts

Tracklists, tracks, search, and most-played tracks are public. Anyone can register
with an email address, a password of 15 to 128 characters, and a display name.
Users must verify their email before creating, updating, or deleting their own
tracklists, or uploading artwork. Tracklist names are unique within each account.
Tracks and artwork remain shared. A public tracklist includes an `owner` with
only `id` and `display_name`; email addresses appear only in `/auth/me` and login
responses.

Set `APP_ORIGIN` to the browser application's origin, such as
`https://memoir.example.com`. It is the allowed CORS origin and the base URL for
account links. Production requires HTTPS on both the browser application and
API. Host them on the same site, such as `memoir.example.com` and
`api.example.com`, or proxy the API through the application's origin. The
`SameSite=Lax` cookies do not support unrelated frontend and API domains.

Local development defaults to `APP_ORIGIN=http://localhost:3000` and an API at
`http://localhost:8080`. Use the same hostname for both. Plain HTTP cookies are
supported only with a localhost origin. `SESSION_TTL` defaults to `168h` and
accepts a positive Go duration up to `720h`.

Sessions use HttpOnly, host-only cookies, with Secure and the `__Host-` prefix
under HTTPS. The database stores token hashes. Logout revokes the current
session, and password reset revokes every session for the account. Passwords use
Argon2id with a random salt, 64 MiB of memory, three iterations, and one lane.

| Endpoint | JSON body | Result |
| --- | --- | --- |
| `GET /auth/csrf` | None | `data.csrf_token` and a CSRF cookie |
| `POST /auth/register` | `email`, `password`, `display_name` | 202 and a generic message |
| `POST /auth/login` | `email`, `password` | `data` containing the account and a session cookie |
| `POST /auth/logout` | None | 204 and a cleared session cookie |
| `GET /auth/me` | None | `data` containing the current account |
| `POST /auth/verify-email` | `token` | 204 |
| `POST /auth/resend-verification` | `email` | 202 and a generic message |
| `POST /auth/forgot-password` | `email` | 202 and a generic message |
| `POST /auth/reset-password` | `token`, `password` | 204; log in again |

Fetch `/auth/csrf` with credentials before posting. Send its token in
`X-CSRF-Token` on every POST, PATCH, or DELETE, including login and logout.
Browser requests must use `credentials: 'include'`. The server requires an exact
matching `Origin` on those requests. JSON endpoints accept `application/json`,
including a charset parameter, and reject unknown fields.

```js
const api = 'http://localhost:8080';
const { data } = await fetch(`${api}/auth/csrf`, {
  credentials: 'include',
}).then(response => response.json());

const response = await fetch(`${api}/auth/login`, {
  method: 'POST',
  credentials: 'include',
  headers: {
    'Content-Type': 'application/json',
    'X-CSRF-Token': data.csrf_token,
  },
  body: JSON.stringify({ email, password }),
});
```

The frontend must implement `/verify-email` and `/reset-password`. Emails link
to these paths with a `#token=...` fragment. Read the fragment, remove it from
browser history, then submit the token in the appropriate API request body.
Verification links expire after 24 hours and reset links after one hour. Both
work once; consuming a link invalidates other outstanding links for the same
account and purpose. A failed delivery does not invalidate an earlier link.

Authentication is limited per process to 50 POST requests per client IP every
15 minutes, with 10 attempts per email and operation in the same period. The
API uses the connection's peer IP and ignores forwarded IP headers. When using
a reverse proxy, apply limits per client at the proxy too. Multiple API
instances need a shared limiter or gateway limits.

Filter public tracklists with `GET /tracklists?user_id=<uuid>`. This combines
with `track_id`, `page`, and `per_page`; page size is capped at 100. Authenticated
writes derive the owner from the session. Missing and non-owned update/delete
targets return 404. `API_TOKEN` and the `API-Token` header no longer authorize
requests. Artwork accepts PNG, JPEG, GIF, and WebP files up to 8 MiB. JSON
content requests are limited to 1 MiB and auth requests to 8 KiB.

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

## Database upgrade

Back up the database and stop the old API before upgrading an existing
installation. The old API cannot run against the final ownership schema.
For an empty database, apply all migrations:

```sh
go tool migrate apply --dsn "$DATABASE_URL" --migrations internal/database/migrations
```

For an existing installation, first apply migrations through the nullable-owner
stage using a temporary directory. Then create the chosen owner's verified
account and assign the old tracklists. The password is read from stdin; do not
put it in command arguments or shell history.

```sh
memoir_migrations=$(mktemp -d)
cp internal/database/migrations/201*.sql "$memoir_migrations/"
cp internal/database/migrations/20261009000100_add_users_and_sessions.sql "$memoir_migrations/"
go tool migrate apply --dsn "$DATABASE_URL" --migrations "$memoir_migrations"

read -rs memoir_password
printf '%s\n' "$memoir_password" | go run ./cmd/memoir-user \
  -email you@example.com -name 'Your name' -claim-legacy-tracklists
unset memoir_password

go tool migrate apply --dsn "$DATABASE_URL" --migrations internal/database/migrations
```

The bootstrap command creates a verified account and claims unowned tracklists
in one transaction. The final migration refuses to run while any unowned rows
remain. Downgrading refuses to restore global name uniqueness if different
owners now share a name; resolve those conflicts first. Configure SMTP and the
browser origin before starting the new API, and update clients to use cookies.

Development CSV seeds should load in this order: `users`, `tracks`, `tracklists`,
`tracklist_tracks`. Track seeds omit the generated search vector. The sample
account has a disabled password hash and cannot
log in. Provision a real account with `memoir-user` instead.

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
