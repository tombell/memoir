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
VALUES ($1, $2, $3, $4, $5, $6, $7)
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
  "tracklists".*,
  "tracks"."id" as "track_id",
  "tracks"."artist",
  "tracks"."name" as "track_name",
  "tracks"."genre",
  "tracks"."bpm",
  "tracks"."key",
  "tracks"."created" as "track_created",
  "tracks"."updated" as "track_updated"
FROM "tracklists"
JOIN "tracklist_tracks" ON "tracklist_tracks"."tracklist_id" = "tracklists"."id"
JOIN "tracks" ON "tracks"."id" = "tracklist_tracks"."track_id"
WHERE "tracklists"."id" = $1
ORDER BY "tracklist_tracks"."track_number" ASC;

-- name: GetTracklists :many
SELECT
  "tracklists"."id",
  "tracklists"."name",
  "tracklists"."date",
  "tracklists"."artwork",
  "tracklists"."url",
  "tracklists"."created",
  "tracklists"."updated",
  count("tracklists"."id") as "track_count"
FROM "tracklists"
JOIN "tracklist_tracks" ON "tracklist_tracks"."tracklist_id" = "tracklists"."id"
GROUP BY "tracklists"."id"
ORDER BY "tracklists"."date" DESC
OFFSET $1
LIMIT $2;

-- name: GetTracklistsByTrack :many
SELECT "tracklists".*, (
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
SET "name" = $2, "url" = $3, "date" = $4, "updated" = now()
WHERE "id" = $1
RETURNING *;
