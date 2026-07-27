INSERT INTO users (email, password, counterparty_id, status)
VALUES ($1, $2, $3, 'active')
RETURNING id
