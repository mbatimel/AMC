UPDATE counterparties
SET name = COALESCE(NULLIF($2, ''), name),
    inn = COALESCE(NULLIF($3, ''), inn),
    updated_at = now()
WHERE id = $1
