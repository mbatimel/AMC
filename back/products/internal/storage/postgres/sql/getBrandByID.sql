SELECT id, name, COALESCE(slug, ''), is_active, created_at, updated_at
FROM brands
WHERE id = $1;
