WITH inserted_file AS (
    INSERT INTO files (storage_key, original_name, mime_type, size_bytes)
    VALUES ($2, $3, $4, $5)
    RETURNING id
)
INSERT INTO product_images (product_id, file_id, url, alt, sort_order, is_main)
SELECT $1, inserted_file.id, $6, $7, $8, $9
FROM inserted_file
RETURNING id, file_id, created_at, updated_at;
