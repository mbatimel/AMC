SELECT
    u.id,
    COALESCE(u.email, ''),
    COALESCE(u.phone, ''),
    COALESCE(u.name, ''),
    COALESCE(u.surename, ''),
    COALESCE(u.middle_name, ''),
    COALESCE(role_info.name, ''),
    COALESCE(u.status, ''),
    client_info.id,
    COALESCE(client_info.name, ''),
    COALESCE(client_info.inn, ''),
    u.is_active,
    u.active_client_id,
    u.created_at,
    u.updated_at,
    u.deleted_at
FROM users u
LEFT JOIN LATERAL (
    SELECT r.name
    FROM user_roles ur
    JOIN roles r ON r.id = ur.role_id
    WHERE ur.user_id = u.id
    ORDER BY r.code
    LIMIT 1
) role_info ON TRUE
LEFT JOIN counterparties client_info
    ON client_info.id = COALESCE(u.active_client_id, u.counterparty_id)
WHERE u.id = $1
  AND u.deleted_at IS NULL
