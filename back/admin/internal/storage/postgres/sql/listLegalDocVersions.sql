SELECT v.id, v.doc_id, v.version, v.summary, v.author, v.created_at,
       v.file_id, COALESCE(f.storage_key, ''), COALESCE(f.original_name, ''),
       COALESCE(f.mime_type, ''), COALESCE(f.size_bytes, 0)
FROM portal_legal_doc_versions v
LEFT JOIN files f ON f.id = v.file_id
WHERE v.doc_id = $1
ORDER BY v.created_at DESC;
