DELETE FROM favorites
WHERE user_id = $1
  AND client_id = $2
  AND product_id = ANY($3::uuid[])
