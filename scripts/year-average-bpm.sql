SELECT avg(t.bpm)::numeric(10,2)
FROM tracklists tl
JOIN tracklist_tracks tt ON tt.tracklist_id = tl.id
JOIN tracks t ON t.id = tt.track_id
WHERE date BETWEEN '2020-01-01' AND '2020-12-31'
