INSERT INTO files (storage_key, original_name, mime_type, size_bytes)
VALUES ($1, $2, $3, $4)
RETURNING id;
