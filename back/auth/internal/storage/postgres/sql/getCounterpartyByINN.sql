SELECT id
FROM counterparties
WHERE inn = $1
LIMIT 1
