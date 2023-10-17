SELECT
  id,
  name,
  date,
  created,
  updated
FROM tracklists
WHERE id = $1
LIMIT 1
