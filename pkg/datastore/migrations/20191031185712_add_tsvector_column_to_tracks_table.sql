-- UP

ALTER TABLE tracks ADD COLUMN tsv TSVECTOR;

UPDATE tracks
SET tsv =
  setweight(to_tsvector(name), 'A') ||
  setweight(to_tsvector(artist), 'B');

CREATE INDEX tracks_tsv_idx ON tracks USING GIN(tsv);

-- DOWN

DROP INDEX IF EXISTS tracks_tsv_idx;

ALTER TABLE tracks DROP tsv;
