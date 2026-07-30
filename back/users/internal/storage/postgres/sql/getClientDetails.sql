SELECT
    c.id,
    COALESCE(c.name, ''),
    COALESCE(c.type, ''),
    COALESCE(c.inn, ''),
    COALESCE(c.ogrn, ''),
    COALESCE(NULLIF(c.actual_address, ''), c.legal_address, addr.address, ''),
    COALESCE(contact.full_name, ''),
    COALESCE(NULLIF(c.phone, ''), contact.phone, ''),
    COALESCE(NULLIF(c.email, ''), contact.email, ''),
    c.created_at,
    c.updated_at
FROM user_clients uc
JOIN users u ON u.id = uc.user_id
JOIN counterparties c ON c.id = uc.client_id
LEFT JOIN LATERAL (
    SELECT a.address
    FROM counterparty_addresses a
    WHERE a.counterparty_id = c.id
    ORDER BY a.is_default DESC, a.id
    LIMIT 1
) addr ON TRUE
LEFT JOIN LATERAL (
    SELECT cc.full_name, cc.phone, cc.email
    FROM counterparty_contacts cc
    WHERE cc.counterparty_id = c.id
    ORDER BY cc.is_primary DESC, cc.id
    LIMIT 1
) contact ON TRUE
WHERE uc.user_id = $1
  AND uc.client_id = $2
  AND u.deleted_at IS NULL
