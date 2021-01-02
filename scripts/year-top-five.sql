SELECT t.artist, t.name, t.genre, count(t.id) as played
FROM tracks t
JOIN tracklist_tracks tt ON tt.track_id = t.id
JOIN tracklists tl ON tl.id = tt.tracklist_id
WHERE tl.date BETWEEN '2020-01-01' AND '2020-12-31'
GROUP BY t.id
ORDER BY played DESC
LIMIT 5
