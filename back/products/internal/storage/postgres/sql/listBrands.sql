SELECT id, name, COALESCE(slug, ''), is_active, created_at, updated_at
FROM brands
ORDER BY name, id
LIMIT $1 OFFSET $2;
