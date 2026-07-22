SELECT id, email, password, status
FROM users
WHERE email = $1
