SELECT
  id,
  artist,
  name,
  genre,
  bpm,
  key,
  created,
  updated
FROM tracks
WHERE artist = $1
AND name = $2
LIMIT 1
