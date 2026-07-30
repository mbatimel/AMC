UPDATE products
SET deleted_at = now(),
    is_active = FALSE,
    updated_at = now()
WHERE id = $1
  AND deleted_at IS NULL;
