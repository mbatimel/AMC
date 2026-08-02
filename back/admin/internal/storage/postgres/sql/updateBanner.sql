UPDATE portal_banners
SET title = $2,
    subtitle = $3,
    image_url = CASE WHEN $10 THEN $4 ELSE image_url END,
    link = $5,
    sort_order = $6,
    is_active = $7,
    date_from = $8,
    date_to = $9,
    file_id = CASE WHEN $10 THEN $11 ELSE file_id END,
    updated_at = now()
WHERE id = $1;
