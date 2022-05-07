-- migrate:up

ALTER TABLE tracks
ADD COLUMN fts_name_and_artist TSVECTOR
GENERATED ALWAYS AS (
  setweight(to_tsvector('english', coalesce(name, '')), 'A')
  || ' ' ||
  setweight(to_tsvector('english', coalesce(artist, '')), 'B')
) STORED;

CREATE INDEX tracks_fts_name_and_artist_idx ON tracks USING GIN(fts_name_and_artist);

-- migrate:down

DROP INDEX IF EXISTS tracks_fts_name_and_artist_idx;

ALTER TABLE tracks DROP fts_name_and_artist;
