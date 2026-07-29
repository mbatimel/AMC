UPDATE users
SET active_client_id = $2,
    counterparty_id = COALESCE(counterparty_id, $2),
    updated_at = now()
WHERE id = $1
  AND deleted_at IS NULL
