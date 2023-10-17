UPDATE tracklists
SET
  name = $2,
  url = $3,
  date = $4,
  updated = NOW()
WHERE id = $1
