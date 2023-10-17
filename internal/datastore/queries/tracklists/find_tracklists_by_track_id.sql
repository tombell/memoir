SELECT tl.*, (
  SELECT count(id)
  FROM tracklist_tracks
  WHERE tracklist_tracks.tracklist_id = tl.id
) as track_count
FROM tracklists tl
JOIN tracklist_tracks tt ON tt.tracklist_id = tl.id
WHERE tt.track_id = $1
ORDER BY tl.date DESC
OFFSET $2 LIMIT $3
