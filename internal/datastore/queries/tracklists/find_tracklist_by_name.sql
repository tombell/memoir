SELECT
  id,
  name,
  date,
  created,
  updated
FROM tracklists
WHERE name = $1
LIMIT 1
