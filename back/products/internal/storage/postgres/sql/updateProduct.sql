UPDATE products
SET
    sku = CASE WHEN $2 THEN $3 ELSE sku END,
    name = CASE WHEN $4 THEN $5 ELSE name END,
    description_full = CASE WHEN $6 THEN $7 ELSE description_full END,
    category_id = CASE WHEN $8 THEN $9 ELSE category_id END,
    brand_id = CASE WHEN $10 THEN $11 ELSE brand_id END,
    gost_tu = CASE WHEN $12 THEN $13 ELSE gost_tu END,
    material = CASE WHEN $14 THEN $15 ELSE material END,
    size = CASE WHEN $16 THEN $17 ELSE size END,
    package_multiple = CASE WHEN $18 THEN $19 ELSE package_multiple END,
    is_active = CASE WHEN $20 THEN $21 ELSE is_active END,
    updated_at = now()
WHERE id = $1
  AND deleted_at IS NULL
RETURNING id;
