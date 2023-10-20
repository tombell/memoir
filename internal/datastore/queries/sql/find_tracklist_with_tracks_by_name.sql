SELECT
  tl.*,
  t.id as track_id,
  t.artist,
  t.name,
  t.genre,
  t.bpm,
  t.key,
  t.created,
  t.updated
FROM tracklists tl
JOIN tracklist_tracks tt ON tt.tracklist_id = tl.id
JOIN tracks t ON t.id = tt.track_id
WHERE tl.name = $1
ORDER BY tt.track_number ASC
