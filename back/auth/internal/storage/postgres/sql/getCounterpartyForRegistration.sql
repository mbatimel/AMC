SELECT
    c.id,
    (
        SELECT COUNT(*)
        FROM counterparties matches
        WHERE regexp_replace(matches.inn, '[[:space:]-]', '', 'g') = $1
    ) AS match_count
FROM counterparties c
WHERE regexp_replace(c.inn, '[[:space:]-]', '', 'g') = $1
ORDER BY c.created_at, c.id
LIMIT 1
FOR UPDATE OF c
