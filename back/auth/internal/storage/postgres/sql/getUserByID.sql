SELECT id, email, password, name, surename, status, created_at, updated_at
FROM users
WHERE id = $1;
