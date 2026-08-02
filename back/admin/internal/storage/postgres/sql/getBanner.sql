SELECT b.id, b.title, b.subtitle, b.image_url, b.link, b.sort_order,
       b.is_active, b.date_from, b.date_to, b.created_at, b.updated_at,
       b.file_id, COALESCE(f.storage_key, ''), COALESCE(f.original_name, ''),
       COALESCE(f.mime_type, ''), COALESCE(f.size_bytes, 0)
FROM portal_banners b
LEFT JOIN files f ON f.id = b.file_id
WHERE b.id = $1;
