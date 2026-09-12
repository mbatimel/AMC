UPDATE products
SET ym_status = $2,
    updated_at = now()
WHERE id = $1
  AND deleted_at IS NULL
RETURNING id;
