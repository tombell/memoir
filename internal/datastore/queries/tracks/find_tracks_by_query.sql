SELECT
  id,
  artist,
  ts_headline(artist, q, 'StartSel=<<, StopSel=>>') as artist_highlighted,
  name,
  ts_headline(name, q, 'StartSel=<<, StopSel=>>') as name_highlighted,
  genre,
  bpm,
  key,
  created,
  updated
FROM (
  SELECT
    id,
    artist,
    name,
    genre,
    bpm,
    key,
    created,
    updated,
    ts_rank(fts_name_and_artist, q) as rank,
    q
  FROM
    tracks,
    websearch_to_tsquery($1) q
  WHERE fts_name_and_artist @@ q
  ORDER BY rank DESC
) as searched_tracks
ORDER BY rank DESC
LIMIT $2
