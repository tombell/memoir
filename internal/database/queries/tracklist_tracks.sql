-- name: AddTracklistTrack :exec
INSERT INTO "tracklist_tracks" (
  "id",
  "tracklist_id",
  "track_id",
  "track_number"
)
VALUES ($1, $2, $3, $4);

-- name: DeleteTracklistTracks :exec
DELETE FROM "tracklist_tracks" WHERE "tracklist_id" = $1;
