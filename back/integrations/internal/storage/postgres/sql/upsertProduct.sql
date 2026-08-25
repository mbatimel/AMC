INSERT INTO products (one_c_guid, category_id, sku, name, is_active)
VALUES ($1, $2, $3, $4, TRUE)
ON CONFLICT (one_c_guid) DO UPDATE
SET category_id = EXCLUDED.category_id,
    sku = EXCLUDED.sku,
    name = EXCLUDED.name,
    updated_at = now()
RETURNING id;
