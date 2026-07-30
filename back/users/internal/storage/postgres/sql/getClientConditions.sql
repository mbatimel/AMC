SELECT
    c.id,
    COALESCE(pg.name, pg.code::text, ''),
    COALESCE(c.credit_limit, 0),
    COALESCE(c.credit_used, 0),
    COALESCE(c.payment_terms, ''),
    COALESCE(c.sales_contact, ''),
    COALESCE(c.contact_channel, '')
FROM user_clients uc
JOIN users u ON u.id = uc.user_id
JOIN counterparties c ON c.id = uc.client_id
LEFT JOIN price_groups pg ON pg.id = c.price_group_id
WHERE uc.user_id = $1
  AND uc.client_id = $2
  AND u.deleted_at IS NULL
