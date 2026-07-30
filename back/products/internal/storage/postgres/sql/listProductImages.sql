SELECT pi.id, pi.product_id, COALESCE(pi.url, f.storage_key, ''), COALESCE(pi.alt, ''),
       COALESCE(pi.sort_order, 0), pi.is_main, pi.created_at, pi.updated_at
FROM product_images pi
LEFT JOIN files f ON f.id = pi.file_id
WHERE pi.product_id = $1
ORDER BY pi.is_main DESC, pi.sort_order, pi.id;
