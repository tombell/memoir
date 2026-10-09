-- name: AddTracklist :one
INSERT INTO tracklists (id, name, url, artwork, date, owner_id, created, updated)
VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
RETURNING *;

-- name: CountTracklists :one
SELECT count(*) FROM tracklists
WHERE (sqlc.narg(owner_id)::uuid IS NULL OR owner_id = sqlc.narg(owner_id))
  AND (sqlc.narg(track_id)::uuid IS NULL OR EXISTS (
    SELECT 1 FROM tracklist_tracks
    WHERE tracklist_tracks.tracklist_id = tracklists.id AND tracklist_tracks.track_id = sqlc.narg(track_id)
  ));

-- name: GetTracklists :many
SELECT sqlc.embed(tracklists), users.display_name AS owner_display_name,
  (SELECT count(*) FROM tracklist_tracks WHERE tracklist_id = tracklists.id) AS track_count
FROM tracklists JOIN users ON users.id = tracklists.owner_id
WHERE (sqlc.narg(owner_id)::uuid IS NULL OR tracklists.owner_id = sqlc.narg(owner_id))
  AND (sqlc.narg(track_id)::uuid IS NULL OR EXISTS (
    SELECT 1 FROM tracklist_tracks
    WHERE tracklist_tracks.tracklist_id = tracklists.id AND tracklist_tracks.track_id = sqlc.narg(track_id)
  ))
ORDER BY tracklists.date DESC, tracklists.id
OFFSET sqlc.arg(page_offset) LIMIT sqlc.arg(page_limit);

-- name: GetTracklistWithTracks :many
SELECT sqlc.embed(tracklists), sqlc.embed(tracks), users.display_name AS owner_display_name
FROM tracklists
JOIN users ON users.id = tracklists.owner_id
JOIN tracklist_tracks ON tracklist_tracks.tracklist_id = tracklists.id
JOIN tracks ON tracks.id = tracklist_tracks.track_id
WHERE tracklists.id = $1
ORDER BY tracklist_tracks.track_number ASC;

-- name: UpdateTracklist :one
UPDATE tracklists
SET name = sqlc.arg(name), url = sqlc.arg(url), date = sqlc.arg(date), updated = NOW()
WHERE id = sqlc.arg(id) AND owner_id = sqlc.arg(owner_id)
RETURNING *;

-- name: LockOwnedTracklist :one
SELECT * FROM tracklists WHERE id = $1 AND owner_id = $2 FOR UPDATE;

-- name: DeleteTracklist :execrows
DELETE FROM tracklists WHERE id = $1 AND owner_id = $2;
