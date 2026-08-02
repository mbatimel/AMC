SELECT b.id, b.title, b.subtitle, b.image_url, b.link, b.sort_order,
       b.is_active, b.date_from, b.date_to, b.created_at, b.updated_at,
       b.file_id, COALESCE(f.storage_key, ''), COALESCE(f.original_name, ''),
       COALESCE(f.mime_type, ''), COALESCE(f.size_bytes, 0)
FROM portal_banners b
LEFT JOIN files f ON f.id = b.file_id
WHERE NOT $1 OR (
    b.is_active = TRUE
    AND (b.date_from IS NULL OR b.date_from <= CURRENT_DATE)
    AND (b.date_to IS NULL OR b.date_to >= CURRENT_DATE)
)
ORDER BY b.sort_order, b.id;
