INSERT INTO favorites (user_id, client_id, product_id)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, client_id, product_id) DO NOTHING
RETURNING user_id, client_id, product_id, created_at
