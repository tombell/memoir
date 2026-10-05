# Memoir

Memoir is a Go HTTP API for DJ mix tracklists. PostgreSQL stores tracklists and
tracks, and AWS S3 stores artwork. The server executable is `cmd/memoir`.

## Local setup

Use Go 1.27 and PostgreSQL. CI uses PostgreSQL 17. Dependencies are vendored, so
building and testing the API does not require downloading modules. SQL generation
uses sqlc 1.30.0, the version recorded in `internal/database/*.go`.

If you use mise, install the versions in `mise.toml`:

```sh
mise install
mise exec -- go version
mise exec -- sqlc version
```

Run the commands below from the repository root with these tools on `PATH`, or
prefix them with `mise exec --`. The API examples also use `curl` and `jq`.

Create a local database. With Docker installed, this starts a PostgreSQL instance
bound to localhost:

```sh
docker run --name memoir-postgres \
  -e POSTGRES_USER=memoir -e POSTGRES_PASSWORD=memoir \
  -e POSTGRES_DB=memoir_development \
  -p 127.0.0.1:5432:5432 -d postgres:17
docker exec memoir-postgres pg_isready -U memoir -d memoir_development
```

Wait until the readiness command reports that PostgreSQL accepts connections.
Alternatively, create `memoir_development` on an existing local PostgreSQL server
and adjust the connection string to match its user and password.

```sh
cp .env.example .env
# Edit .env for your local database and development S3 bucket.
set -a
. ./.env
set +a

go tool migrate apply --db postgresql --dsn "$DATABASE_URL" \
  --migrations internal/database/migrations
make dev
./bin/memoir
```

Keep the server running while using the examples below. The server loads `.env`
from its working directory. Existing environment variables take precedence.
The migration tool does not load `.env`, so export `DATABASE_URL` before invoking
it. No separate `migrate` executable or `go install` is needed.

## Configuration

| Variable | Default or requirement | Purpose |
| --- | --- | --- |
| `HOST` | `127.0.0.1` | HTTP listen address |
| `PORT` | `8080` | HTTP listen port |
| `DATABASE_URL` | Required | PostgreSQL URL or pgx connection string |
| `API_TOKEN` | Required | Shared token for write requests |
| `AWS_BUCKET` | Required | S3 bucket for artwork |
| `AWS_REGION` | Required | Region containing the bucket |
| `AWS_KEY` | Required | Static AWS access key ID |
| `AWS_SECRET` | Required | Static AWS secret access key |

All required values must be nonempty, including the AWS values when using only
database endpoints. `.env.example` supplies dummy AWS credentials for local
database work. Replace them before uploading artwork. The current server uses
static credentials from these variables.

Use a development bucket with `s3:GetObject` and `s3:PutObject` access for its
objects. Create the bucket separately. Memoir returns object keys and does not
serve artwork or make objects publicly readable. Configure any public artwork
URL or CDN outside the API. Keep real credentials in the ignored `.env` file or
the environment.

## Migrations and SQL generation

Migrations are ordered SQL files in `internal/database/migrations`, with
`-- migrate:up` and `-- migrate:down` sections. `apply` skips migrations already
recorded in the database. To roll back the latest migration on a local database:

```sh
go tool migrate rollback --db postgresql --dsn "$DATABASE_URL" \
  --migrations internal/database/migrations --steps 1
```

Rollback commands can remove data. Use a disposable local database when testing
them.

Edit SQL in `internal/database/queries` and migrations, then regenerate and
verify the Go code:

```sh
sqlc version # must print v1.30.0
make sqlc
make check
```

`make sqlc` rejects other generator versions. `sqlc.yaml` defines the schema,
queries, and output package. Commit the generated Go changes with the SQL changes;
do not edit the generated files or `vendor` by hand.

The CSV files in `internal/database/seed` are historical exports. In particular,
`tracks.csv` includes the generated `fts_name_and_artist` column, which PostgreSQL
does not accept as an imported value. Do not copy that file directly into the
current table. The API examples below populate a development database without
importing those exports.

## API

Reads are public. `POST`, `PATCH`, and `DELETE` routes require the configured
token in the `API-Token` header. A missing token returns 401, and an incorrect token
returns 403. JSON writes should send `Content-Type: application/json` exactly.

Successful JSON responses wrap results in `data`. The tracklist index also
returns `meta.current_page` and `meta.total_pages`. Tracklists expose `trackCount`;
tracks expose `bpm`, `key`, and optional `played`, `artistHighlighted`, and
`nameHighlighted` fields. Search highlights use `<<` and `>>` markers.

| Method | Path | Input |
| --- | --- | --- |
| `GET` | `/tracklists` | `page`, `per_page`, optional `track_id` |
| `GET` | `/tracklists/{id}` | Tracklist UUID |
| `POST` | `/tracklists` | JSON tracklist with tracks |
| `PATCH` | `/tracklists/{id}` | JSON `name`, `date`, and `url` |
| `DELETE` | `/tracklists/{id}` | Tracklist UUID |
| `GET` | `/tracks/{id}` | Track UUID |
| `GET` | `/tracks/search` | `q`, optional `per_page` |
| `GET` | `/tracks/mostplayed` | Optional `per_page` |
| `POST` | `/artwork` | Multipart file field named `artwork` |

Tracklist pagination defaults to page 1 and 10 results per page. Search and
most-played currently return at most `per_page` results, defaulting to 10; their
`page` parameter is not implemented on this branch. There is no general
`GET /tracks` endpoint or direct track mutation endpoint.

### Tracklist input

Creation accepts `name`, `date`, `url`, `artwork`, and `tracks`. Use a unique mix
name, an RFC 3339 timestamp such as `2026-10-05T18:00:00Z`, a URL, and an artwork
object key. Supply at least one track. Each track is an array of five strings in
this exact order:

```json
["Track name", "Artist", "125", "1A", "House"]
```

The fields are **name, artist, BPM, key, genre**. BPM is a numeric string, and the
server uppercases the musical key. The outer array order defines track numbers.
Creation reuses tracks with the same artist and name. Read responses return
track objects rather than these input arrays.

On this branch, `PATCH` requires all three of `name`, `date`, and `url`. It updates
mix metadata; it does not replace tracks or artwork. Fetch the tracklist after
creating it to read its tracks and track count.

### Runnable examples

In a second shell, export the same local configuration and set the API URL:

```sh
set -a
. ./.env
set +a
BASE_URL=http://127.0.0.1:8080

curl --fail-with-body "$BASE_URL/tracklists?page=1&per_page=10" | jq .
```

Create a sample mix. `example.png` is a sample artwork key; replace it with the
key returned by an upload if you want the mix to refer to an actual S3 object.
Choose a different name when repeating creation before deleting the sample.

```sh
created=$(curl --fail-with-body -sS "$BASE_URL/tracklists" \
  -H "API-Token: $API_TOKEN" -H 'Content-Type: application/json' \
  --data-binary @- <<'JSON'
{
  "name": "Local development mix",
  "date": "2026-10-05T18:00:00Z",
  "url": "https://example.com/mixes/local",
  "artwork": "example.png",
  "tracks": [
    ["Tonight", "Example Artist", "125", "1A", "House"],
    ["After Hours", "Another Artist", "126", "2A", "Tech House"]
  ]
}
JSON
)
printf '%s\n' "$created" | jq .
TRACKLIST_ID=$(printf '%s\n' "$created" | jq -er '.data.id')

mix=$(curl --fail-with-body -sS "$BASE_URL/tracklists/$TRACKLIST_ID")
printf '%s\n' "$mix" | jq .
TRACK_ID=$(printf '%s\n' "$mix" | jq -er '.data.tracks[0].id')

curl --fail-with-body "$BASE_URL/tracks/$TRACK_ID" | jq .
curl --fail-with-body "$BASE_URL/tracklists?track_id=$TRACK_ID" | jq .
curl --fail-with-body --get "$BASE_URL/tracks/search" \
  --data-urlencode 'q=Tonight' --data-urlencode 'per_page=10' | jq .
curl --fail-with-body "$BASE_URL/tracks/mostplayed?per_page=10" | jq .
```

Update the mix and then delete it:

```sh
curl --fail-with-body -X PATCH "$BASE_URL/tracklists/$TRACKLIST_ID" \
  -H "API-Token: $API_TOKEN" -H 'Content-Type: application/json' \
  --data '{"name":"Updated local mix","date":"2026-10-05T19:00:00Z","url":"https://example.com/mixes/updated"}' | jq .

curl --fail-with-body -i -X DELETE "$BASE_URL/tracklists/$TRACKLIST_ID" \
  -H "API-Token: $API_TOKEN"
```

The delete route returns 204. Deleting a tracklist removes its track associations;
tracks can remain available independently.

### Artwork upload

With credentials for a development S3 bucket, upload an existing image file:

```sh
curl --fail-with-body "$BASE_URL/artwork" \
  -H "API-Token: $API_TOKEN" -F 'artwork=@./cover.png' | jq .
```

Let curl set the multipart `Content-Type` and boundary. The response has the form
`{"data":{"key":"<object-key>"}}`. Use that key in the `artwork` field when
creating a tracklist. The current key combines the file's MD5 digest and filename
extension. Upload returns 201 for a new object or 200 when that key already exists.
There is no artwork download or delete route.

## Verification

```sh
make fmt   # format project Go files, excluding vendor
make check # check formatting, build, vet, and run unit tests
```

The individual targets are `check-format`, `build`, `vet`, and `test`. Formatting
checks fail without rewriting files. Build, vet, and tests use vendored modules.
Tests under `scripts` cover the verification commands and do not contact a
database or S3.

For PostgreSQL integration tests, provide a dedicated disposable test database:

```sh
export MEMOIR_TEST_DATABASE_URL='postgresql://memoir:memoir@127.0.0.1:5432/memoir_test?sslmode=disable'
# Create memoir_test on your local PostgreSQL server before running this.
make test-integration
```

The target requires `MEMOIR_TEST_DATABASE_URL`, applies the migrations to that
database, overrides `DATABASE_URL` with it, and runs
`go test -mod=vendor -tags=integration -count=1 ./...`. Integration tests should use
`MEMOIR_TEST_DATABASE_URL` and may use the `integration` build tag. Tests without
the tag must skip database work when this variable is unset. Tests may alter the
database; do not use an application database. The target does not create or drop
the database.

GitHub Actions runs `make check` and a separate integration job with a fresh
PostgreSQL 17 service and explicit test-only credentials, using GitHub's
[PostgreSQL service configuration](https://docs.github.com/en/actions/tutorials/use-containerized-services/create-postgresql-service-containers).
It applies migrations
even when no database tests are present, and includes integration tests as they
are added. Neither job uses production secrets or a live S3 bucket. This branch
contains verification-command tests, not database-backed API tests.

## Container build

```sh
docker build -t memoir .
```

The builder uses Go 1.27 and the vendored dependencies. The runtime image contains
the binary and CA certificates and runs as an unprivileged user. `.dockerignore`
excludes `.env`. Pass configuration at runtime, set `HOST=0.0.0.0` to accept
connections through a published port, and use a database address reachable from
inside the container. Migrations run separately before starting the API.
