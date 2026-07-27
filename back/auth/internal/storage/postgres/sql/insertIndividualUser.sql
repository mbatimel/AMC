INSERT INTO users (email, password, surename, name, middle_name, phone, city, delivery_address, inn, status)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, 'active')
RETURNING id
