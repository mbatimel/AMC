INSERT INTO products (
    sku,
    name,
    description_full,
    category_id,
    brand_id,
    gost_tu,
    material,
    size,
    package_multiple,
    is_active
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING id;
