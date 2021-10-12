-- UP

CREATE TABLE tracklists (
  id      UUID PRIMARY KEY,
  name    VARCHAR(256) UNIQUE NOT NULL,
  date    TIMESTAMP NOT NULL,
  artwork VARCHAR(256) NOT NULL,
  url     VARCHAR(256) NOT NULL,
  created TIMESTAMP NOT NULL,
  updated TIMESTAMP NOT NULL
);

CREATE INDEX tracklists_name_idx ON tracklists ((lower(name)));

-- DOWN

DROP INDEX IF EXISTS tracklists_name_idx;

DROP TABLE IF EXISTS tracklists;
