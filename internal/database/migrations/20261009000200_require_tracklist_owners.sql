-- migrate:up

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM tracklists WHERE owner_id IS NULL) THEN
    RAISE EXCEPTION 'Unowned tracklists remain. Provision an owner with memoir-user -claim-legacy-tracklists before this migration.';
  END IF;
END $$;

ALTER TABLE tracklists ALTER COLUMN owner_id SET NOT NULL;
ALTER TABLE tracklists DROP CONSTRAINT tracklists_name_key;
ALTER TABLE tracklists ADD CONSTRAINT tracklists_owner_name_key UNIQUE (owner_id, name);

-- migrate:down

-- This deliberately fails if different owners now share a name. Resolve those
-- conflicts before downgrading rather than silently deleting or renaming data.
ALTER TABLE tracklists ADD CONSTRAINT tracklists_name_key UNIQUE (name);
ALTER TABLE tracklists DROP CONSTRAINT tracklists_owner_name_key;
ALTER TABLE tracklists ALTER COLUMN owner_id DROP NOT NULL;
