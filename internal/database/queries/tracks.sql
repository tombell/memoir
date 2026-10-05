-- name: AddTrack :exec
INSERT INTO "tracks" (
  "id",
  "artist",
  "name",
  "genre",
  "bpm",
  "key",
  "created",
  "updated"
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: CountMostPlayedTracks :one
SELECT count(DISTINCT "track_id") FROM "tracklist_tracks";

-- name: GetMostPlayedTracks :many
SELECT
  sqlc.embed(tracks),
  count(DISTINCT "tracklist_tracks"."tracklist_id") as "played"
FROM "tracks"
JOIN "tracklist_tracks" ON "tracklist_tracks"."track_id" = "tracks"."id"
GROUP BY "tracks"."id"
ORDER BY "played" DESC, "tracks"."id" ASC
LIMIT sqlc.arg(row_limit)::int
OFFSET sqlc.arg(row_offset)::int;

-- name: GetTrack :one
SELECT
  "id",
  "artist",
  "name",
  "genre",
  "bpm",
  "key",
  "created",
  "updated"
FROM "tracks"
WHERE "id" = $1
LIMIT 1;

-- name: GetTrackByArtistAndName :one
SELECT
  "id",
  "artist",
  "name",
  "genre",
  "bpm",
  "key",
  "created",
  "updated"
FROM "tracks"
WHERE "artist" = $1 AND "name" = $2
LIMIT 1;

-- name: CountTracksByQuery :one
SELECT count("id")
FROM "tracks", websearch_to_tsquery('english', sqlc.arg(query)::text) "q"
WHERE "fts_name_and_artist" @@ "q";

-- name: GetTracksByQuery :many
SELECT
  "id",
  "artist",
  ts_headline("artist", "q", 'StartSel=<<, StopSel=>>') as "artist_highlighted",
  "name",
  ts_headline("name", "q", 'StartSel=<<, StopSel=>>') as "name_highlighted",
  "genre",
  "bpm",
  "key",
  "created",
  "updated"
FROM (
  SELECT
    "id",
    "artist",
    "name",
    "genre",
    "bpm",
    "key",
    "created",
    "updated",
    ts_rank("fts_name_and_artist", "q") as "rank",
    "q"
  FROM "tracks", websearch_to_tsquery('english', sqlc.arg(query)::text) "q"
  WHERE "fts_name_and_artist" @@ "q"
) as "searched_tracks"
ORDER BY "rank" DESC, "id" ASC
LIMIT sqlc.arg(row_limit)::int
OFFSET sqlc.arg(row_offset)::int;
