-- UP

ALTER TABLE tracklists
ADD artwork VARCHAR(256) AFTER name;

-- DOWN

ALTER TABLE tracklists
DROP artwork;
