SELECT c.id, c.title, c.sort_order, c.is_active, c.created_at, c.updated_at,
       c.file_id, COALESCE(f.storage_key, ''), COALESCE(f.original_name, ''),
       COALESCE(f.mime_type, ''), COALESCE(f.size_bytes, 0)
FROM portal_certificates c
LEFT JOIN files f ON f.id = c.file_id
WHERE c.id = $1;
