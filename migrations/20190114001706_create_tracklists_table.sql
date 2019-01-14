-- UP

CREATE TABLE tracklists (
  id uuid PRIMARY KEY
);

-- DOWN

DROP TABLE IF EXISTS tracklists CASCADE;
