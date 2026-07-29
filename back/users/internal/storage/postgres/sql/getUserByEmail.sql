SELECT id
FROM users
WHERE LOWER(email) = LOWER($1)
  AND deleted_at IS NULL
