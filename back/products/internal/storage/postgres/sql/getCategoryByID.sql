SELECT id, COALESCE(name, ''), COALESCE(slug, ''), parent_id,
       COALESCE(sort_order, 0), is_active, created_at, updated_at
FROM categories
WHERE id = $1;
