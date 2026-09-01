INSERT INTO users (
    email,
    password,
    counterparty_id,
    active_client_id,
    surename,
    name,
    middle_name,
    phone,
    status
)
VALUES ($1, $2, $3, $3, NULLIF($4, ''), NULLIF($5, ''), NULLIF($6, ''), $7, 'active')
RETURNING id
