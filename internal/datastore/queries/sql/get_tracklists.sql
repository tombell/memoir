SELECT tl.*, count(tl.id) as track_count
FROM tracklists tl
JOIN tracklist_tracks tt ON tt.tracklist_id = tl.id
GROUP BY tl.id
ORDER BY tl.date DESC
OFFSET $1 LIMIT $2
