-- UP

CREATE TABLE tracklist_tracks (
  id           UUID PRIMARY KEY,
  tracklist_id UUID REFERENCES tracklists (id),
  track_id     UUID REFERENCES tracks (id)
);

-- DOWN

DROP TABLE IF EXISTS tracklist_tracks CASCADE;
