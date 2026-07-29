INSERT INTO user_clients (user_id, client_id, is_default)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, client_id) DO UPDATE
SET updated_at = now()
