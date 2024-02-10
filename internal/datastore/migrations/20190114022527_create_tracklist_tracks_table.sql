-- migrate:up

CREATE TABLE "tracklist_tracks" (
  "id"           UUID    PRIMARY KEY,
  "tracklist_id" UUID    REFERENCES "tracklists" ("id") NOT NULL,
  "track_id"     UUID    REFERENCES "tracks" ("id") NOT NULL,
  "track_number" INTEGER NOT NULL
);

CREATE INDEX "tracklist_tracks_tracklist_id_idx" ON "tracklist_tracks" ("tracklist_id");
CREATE INDEX "tracklist_tracks_track_id_idx" ON "tracklist_tracks" ("track_id");

-- migrate:down

DROP INDEX IF EXISTS "tracklist_tracks_track_id_idx";
DROP INDEX IF EXISTS "tracklist_tracks_tracklist_id_idx";

DROP TABLE IF EXISTS "tracklist_tracks";
