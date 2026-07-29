SELECT user_id, client_id, product_id, created_at
FROM favorites
WHERE user_id = $1
  AND client_id = $2
ORDER BY created_at DESC, product_id
