INSERT INTO users (email, password, name, surename, status)
VALUES ($1, $2, $3, $4, 'active')
RETURNING id
