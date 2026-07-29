UPDATE users
SET email = COALESCE(NULLIF($2, ''), email),
    phone = COALESCE(NULLIF($3, ''), phone),
    name = COALESCE(NULLIF($4, ''), name),
    surename = COALESCE(NULLIF($5, ''), surename),
    middle_name = CASE WHEN $7 THEN $6 ELSE middle_name END,
    updated_at = now()
WHERE id = $1
  AND deleted_at IS NULL
RETURNING id
