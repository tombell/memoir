-- UP

CREATE TABLE tracklists (
  id      UUID PRIMARY KEY,
  created TIMESTAMP NOT NULL,
  updated TIMESTAMP NOT NULL
);

-- DOWN

DROP TABLE IF EXISTS tracklists CASCADE;
