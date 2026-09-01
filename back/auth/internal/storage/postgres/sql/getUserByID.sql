SELECT id, email, password, status
FROM users
WHERE id = $1
  AND deleted_at IS NULL
  AND is_active = TRUE
