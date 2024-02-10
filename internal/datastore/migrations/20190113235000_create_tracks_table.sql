-- migrate:up

CREATE TABLE "tracks" (
  "id"      UUID          PRIMARY KEY,
  "artist"  VARCHAR (256) NOT NULL,
  "name"    VARCHAR (256) NOT NULL,
  "genre"   VARCHAR (128) NOT NULL,
  "bpm"     DECIMAL       NOT NULL,
  "key"     VARCHAR (8)   NOT NULL,
  "created" TIMESTAMP     NOT NULL,
  "updated" TIMESTAMP     NOT NULL
);

CREATE INDEX "tracks_artist_idx" ON "tracks" (lower("artist"));
CREATE INDEX "tracks_name_idx"   ON "tracks" (lower("name"));
CREATE INDEX "tracks_genre_idx"  ON "tracks" (lower("genre"));

-- migrate:down

DROP INDEX IF EXISTS "tracks_genre_idx";
DROP INDEX IF EXISTS "tracks_name_idx";
DROP INDEX IF EXISTS "tracks_artist_idx";

DROP TABLE IF EXISTS "tracks";
