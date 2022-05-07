-- name: AddTracklist :exec
INSERT INTO tracklists (id, name, url, artwork, date, created, updated)
    VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: UpdateTracklist :exec
UPDATE
    tracklists
SET
    name = $2,
    url = $3,
    date = $4,
    updated = NOW()
WHERE
    id = $1;

-- name: GetTracklistsCount :one
SELECT
    COUNT(id)
FROM
    tracklists;

-- name: GetTracklists :many
SELECT
    tl.*,
    count(tl.id) AS track_count
FROM
    tracklists tl
    JOIN tracklist_tracks tt ON tt.tracklist_id = tl.id
GROUP BY
    tl.id
ORDER BY
    tl.date DESC OFFSET $1
LIMIT $2;

-- name: FindTracklistByID :one
SELECT
    id,
    name,
    date,
    created,
    updated
FROM
    tracklists
WHERE
    id = $1
LIMIT 1;

-- name: FindTracklistWithTracksByID :many
SELECT
    tl.*,
    t.id AS track_id,
    t.artist,
    t.name,
    t.genre,
    t.bpm,
    t.key,
    t.created,
    t.updated
FROM
    tracklists tl
    JOIN tracklist_tracks tt ON tt.tracklist_id = tl.id
    JOIN tracks t ON t.id = tt.track_id
WHERE
    tl.id = $1
ORDER BY
    tt.track_number ASC;

-- name: FindTracklistByName :one
SELECT
    id,
    name,
    date,
    created,
    updated
FROM
    tracklists
WHERE
    name = $1
LIMIT 1;

-- name: FindTracklistWithTracksByName :one
SELECT
    tl.*,
    t.id AS track_id,
    t.artist,
    t.name,
    t.genre,
    t.bpm,
    t.key,
    t.created,
    t.updated
FROM
    tracklists tl
    JOIN tracklist_tracks tt ON tt.tracklist_id = tl.id
    JOIN tracks t ON t.id = tt.track_id
WHERE
    tl.name = $1
ORDER BY
    tt.track_number ASC;

-- name: FindTracklistsByTrackIDCount :one
SELECT
    COUNT(tracklists.id)
FROM (
    SELECT
        tl.id
    FROM
        tracklists tl
        JOIN tracklist_tracks tt ON tt.tracklist_id = tl.id
    WHERE
        tt.track_id = $1
    GROUP BY
        tl.id
    ORDER BY
        tl.date DESC) AS tracklists;

-- name: FindTracklistsByTrackID :many
SELECT
    tl.*,
    (
        SELECT
            count(id)
        FROM
            tracklist_tracks
        WHERE
            tracklist_tracks.tracklist_id = tl.id) AS track_count
FROM
    tracklists tl
    JOIN tracklist_tracks tt ON tt.tracklist_id = tl.id
WHERE
    tt.track_id = $1
ORDER BY
    tl.date DESC OFFSET $2
LIMIT $3;

