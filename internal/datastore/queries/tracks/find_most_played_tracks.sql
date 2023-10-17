SELECT
  t.id,
  t.artist,
  t.name,
  t.genre,
  t.bpm,
  t.key,
  t.created,
  t.updated,
  count(t.id) as played
FROM tracks t
JOIN tracklist_tracks tt ON tt.track_id = t.id
GROUP BY t.id
ORDER BY played DESC
LIMIT $1
