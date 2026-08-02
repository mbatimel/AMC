SELECT pi.id, pi.product_id, COALESCE(pi.file_id, '00000000-0000-0000-0000-000000000000'::UUID), COALESCE(pi.url, ''),
       COALESCE(f.storage_key, ''), COALESCE(f.original_name, ''),
       COALESCE(f.mime_type, ''), COALESCE(f.size_bytes, 0),
       COALESCE(pi.alt, ''), COALESCE(pi.sort_order, 0), pi.is_main,
       pi.created_at, pi.updated_at
FROM product_images pi
LEFT JOIN files f ON f.id = pi.file_id
WHERE pi.product_id = $1 AND pi.id = $2;
