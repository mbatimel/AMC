INSERT INTO portal_banners (
    id, title, subtitle, image_url, link, sort_order, is_active,
    date_from, date_to, file_id
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);
