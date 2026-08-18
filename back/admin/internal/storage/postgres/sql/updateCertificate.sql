UPDATE portal_certificates
SET title = $2,
    sort_order = $3,
    is_active = $4,
    file_id = CASE WHEN $5 THEN $6 ELSE file_id END,
    updated_at = now()
WHERE id = $1;
