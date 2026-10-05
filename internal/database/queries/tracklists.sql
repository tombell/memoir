-- name: AddTracklist :one
INSERT INTO "tracklists" (
  "id",
  "name",
  "url",
  "artwork",
  "date",
  "created",
  "updated"
)
VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
RETURNING *;

-- name: CountTracklists :one
SELECT count("id") FROM "tracklists";

-- name: CountTracklistsByTrack :one
SELECT
  count("tracklists"."id")
FROM (
  SELECT "tracklists"."id"
  FROM "tracklists"
  JOIN "tracklist_tracks" ON "tracklist_tracks"."tracklist_id" = "tracklists"."id"
  WHERE "tracklist_tracks"."track_id" = $1
  GROUP BY "tracklists"."id"
  ORDER BY "tracklists"."date" DESC
) AS "tracklists";

-- name: GetTracklistWithTracks :many
SELECT
  sqlc.embed(tracklists),
  sqlc.embed(tracks)
FROM "tracklists"
JOIN "tracklist_tracks" ON "tracklist_tracks"."tracklist_id" = "tracklists"."id"
JOIN "tracks" ON "tracks"."id" = "tracklist_tracks"."track_id"
WHERE "tracklists"."id" = $1
ORDER BY "tracklist_tracks"."track_number" ASC;

-- name: GetTracklists :many
SELECT
  sqlc.embed(tracklists),
  count("tracklists"."id") as "track_count"
FROM "tracklists"
JOIN "tracklist_tracks" ON "tracklist_tracks"."tracklist_id" = "tracklists"."id"
GROUP BY "tracklists"."id"
ORDER BY "tracklists"."date" DESC
OFFSET $1
LIMIT $2;

-- name: GetTracklistsByTrack :many
SELECT
  sqlc.embed(tracklists),
  (
    SELECT count("id")
    FROM "tracklist_tracks"
    WHERE "tracklist_tracks"."tracklist_id" = "tracklists"."id"
  ) as "track_count"
FROM "tracklists"
JOIN "tracklist_tracks" ON "tracklist_tracks"."tracklist_id" = "tracklists"."id"
WHERE "tracklist_tracks"."track_id" = $1
ORDER BY "tracklists"."date" DESC
OFFSET $2
LIMIT $3;

-- name: UpdateTracklist :one
UPDATE "tracklists"
SET
  "name" = CASE WHEN sqlc.arg(update_name)::boolean THEN sqlc.arg(name)::text ELSE "name" END,
  "url" = CASE WHEN sqlc.arg(update_url)::boolean THEN sqlc.arg(url)::text ELSE "url" END,
  "date" = CASE WHEN sqlc.arg(update_date)::boolean THEN sqlc.arg(date)::timestamp ELSE "date" END,
  "artwork" = CASE WHEN sqlc.arg(update_artwork)::boolean THEN sqlc.arg(artwork)::text ELSE "artwork" END,
  "updated" = NOW()
WHERE "id" = sqlc.arg(id)
RETURNING *;

-- name: GetTracklist :one
SELECT * FROM "tracklists" WHERE "id" = $1;

-- name: DeleteTracklist :exec
DELETE FROM "tracklists" WHERE "id" = $1;
