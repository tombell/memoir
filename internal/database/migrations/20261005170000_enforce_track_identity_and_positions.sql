-- migrate:up

-- The migration runner executes this migration in a transaction. Hold the
-- tables locked so writes cannot introduce conflicts between the audit and DDL.
LOCK TABLE "tracks", "tracklist_tracks" IN ACCESS EXCLUSIVE MODE;

DO $audit$
DECLARE
  conflict RECORD;
BEGIN
  SELECT "artist", "name", array_agg("id" ORDER BY "id") AS ids
  INTO conflict
  FROM "tracks"
  GROUP BY "artist", "name"
  HAVING count(*) > 1
  ORDER BY "artist", "name"
  LIMIT 1;

  IF FOUND THEN
    RAISE EXCEPTION USING
      MESSAGE = 'Cannot enforce track identity uniqueness: duplicate artist/name identities exist.',
      DETAIL = format('artist=%L, name=%L, track IDs=%s', conflict.artist, conflict.name, conflict.ids),
      HINT = 'Audit with SELECT artist, name, array_agg(id ORDER BY id) AS track_ids FROM tracks GROUP BY artist, name HAVING count(*) > 1; resolve each identity and its tracklist_tracks references explicitly, then retry. No rows have been merged or deleted.';
  END IF;

  SELECT "id", "tracklist_id", "track_number"
  INTO conflict
  FROM "tracklist_tracks"
  WHERE "track_number" <= 0
  ORDER BY "id"
  LIMIT 1;

  IF FOUND THEN
    RAISE EXCEPTION USING
      MESSAGE = 'Cannot enforce positive track positions: nonpositive positions exist.',
      DETAIL = format('row ID=%s, tracklist ID=%s, position=%s', conflict.id, conflict.tracklist_id, conflict.track_number),
      HINT = 'Audit with SELECT * FROM tracklist_tracks WHERE track_number <= 0; correct the positions explicitly, then retry.';
  END IF;

  SELECT "tracklist_id", "track_number", array_agg("id" ORDER BY "id") AS ids
  INTO conflict
  FROM "tracklist_tracks"
  GROUP BY "tracklist_id", "track_number"
  HAVING count(*) > 1
  ORDER BY "tracklist_id", "track_number"
  LIMIT 1;

  IF FOUND THEN
    RAISE EXCEPTION USING
      MESSAGE = 'Cannot enforce unique track positions: duplicate positions exist within a tracklist.',
      DETAIL = format('tracklist ID=%s, position=%s, row IDs=%s', conflict.tracklist_id, conflict.track_number, conflict.ids),
      HINT = 'Audit with SELECT tracklist_id, track_number, array_agg(id ORDER BY id) AS row_ids FROM tracklist_tracks GROUP BY tracklist_id, track_number HAVING count(*) > 1; correct the positions explicitly, then retry.';
  END IF;
END;
$audit$;

ALTER TABLE "tracks"
  ADD CONSTRAINT "tracks_artist_name_key" UNIQUE ("artist", "name");

ALTER TABLE "tracklist_tracks"
  ADD CONSTRAINT "tracklist_tracks_positive_position" CHECK ("track_number" > 0),
  ADD CONSTRAINT "tracklist_tracks_tracklist_position_key" UNIQUE ("tracklist_id", "track_number");

-- migrate:down

ALTER TABLE "tracklist_tracks"
  DROP CONSTRAINT "tracklist_tracks_tracklist_position_key",
  DROP CONSTRAINT "tracklist_tracks_positive_position";

ALTER TABLE "tracks" DROP CONSTRAINT "tracks_artist_name_key";
