SELECT c.id, COALESCE(c.name, ''), COALESCE(c.slug, ''), c.parent_id,
       COALESCE(c.sort_order, 0), c.is_active, c.created_at, c.updated_at,
       COALESCE(pc.items_count, 0)
FROM categories c
LEFT JOIN LATERAL (
    SELECT COUNT(*) AS items_count
    FROM products p
    WHERE p.category_id = c.id
) pc ON TRUE
WHERE c.id = $1;
