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
WHERE id = $1
LIMIT 1
