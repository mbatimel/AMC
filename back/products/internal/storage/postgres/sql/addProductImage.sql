INSERT INTO product_images (product_id, url, alt, sort_order, is_main)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, created_at, updated_at;
