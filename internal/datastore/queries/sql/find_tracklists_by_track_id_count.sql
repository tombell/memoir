SELECT
  COUNT(tracklists.id)
FROM (
  SELECT tl.id
  FROM tracklists tl
  JOIN tracklist_tracks tt ON tt.tracklist_id = tl.id
  WHERE tt.track_id = $1
  GROUP BY tl.id
  ORDER BY tl.date DESC
) AS tracklists
