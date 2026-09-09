INSERT INTO users (email, password, status)
VALUES ($1, $2, 'active')
RETURNING id
