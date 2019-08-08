-- UP

ALTER TABLE tracklists
ADD artwork VARCHAR(256);

-- DOWN

ALTER TABLE tracklists
DROP artwork;
