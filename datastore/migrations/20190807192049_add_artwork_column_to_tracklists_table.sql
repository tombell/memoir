-- UP

ALTER TABLE tracklists
ADD artwork VARCHAR(256) NOT NULL DEFAULT '';

-- DOWN

ALTER TABLE tracklists
DROP artwork;
