SELECT d.id, d.name, d.body, d.current_version, d.updated_at,
       d.current_file_id, COALESCE(f.storage_key, ''), COALESCE(f.original_name, ''),
       COALESCE(f.mime_type, ''), COALESCE(f.size_bytes, 0)
FROM portal_legal_docs d
LEFT JOIN files f ON f.id = d.current_file_id
WHERE d.id = $1;
