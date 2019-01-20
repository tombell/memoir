-- UP

CREATE TABLE mix_uploads (
  id           UUID PRIMARY KEY,
  tracklist_id UUID REFERENCES tracklists (id) ON DELETE CASCADE,
  filename     VARCHAR(256) UNIQUE NOT NULL,
  location     VARCHAR(256) UNIQUE NOT NULL,
  created      TIMESTAMP NOT NULL,
  updated      TIMESTAMP NOT NULL
);

CREATE INDEX mix_uploads_tracklist_id_idx ON mix_uploads (tracklist_id);

-- DOWN

DROP INDEX IF EXISTS mix_uploads_tracklist_id_idx CASCADE;

DROP TABLE IF EXISTS mix_uploads CASCADE;
