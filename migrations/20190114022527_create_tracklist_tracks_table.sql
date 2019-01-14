-- UP

CREATE TABLE tracklist_tracks (
  id           uuid PRIMARY KEY,
  tracklist_id uuid REFERENCES tracklists (id),
  track_id     uuid REFERENCES tracks (id)
);

-- DOWN

DROP TABLE IF EXISTS tracklist_tracks CASCADE;
