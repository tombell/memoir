-- UP

CREATE TABLE tracks (
  id     UUID PRIMARY KEY,
  artist VARCHAR (256) NOT NULL,
  name   VARCHAR (256) NOT NULL,
  genre  VARCHAR (128) NOT NULL,
  bpm    INTEGER NOT NULL,
  key    VARCHAR (8) NOT NULL
);

CREATE INDEX tracks_artist_idx ON tracks ((lower(artist)));
CREATE INDEX tracks_name_idx ON tracks ((lower(name)));
CREATE INDEX tracks_genre_idx ON tracks ((lower(genre)));

-- DOWN

DROP INDEX IF EXISTS tracks_genre_idx CASCADE;
DROP INDEX IF EXISTS tracks_name_idx CASCADE;
DROP INDEX IF EXISTS tracks_artist_idx CASCADE;

DROP TABLE IF EXISTS tracks CASCADE;
