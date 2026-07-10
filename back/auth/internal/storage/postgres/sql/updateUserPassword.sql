UPDATE users
SET password = $2, updated_at = now()
WHERE id = $1;
