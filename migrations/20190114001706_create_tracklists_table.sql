-- UP

CREATE TABLE tracklists (
  id      uuid PRIMARY KEY,
  created TIMESTAMP NOT NULL
);

-- DOWN

DROP TABLE IF EXISTS tracklists CASCADE;
