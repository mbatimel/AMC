UPDATE users
SET counterparty_id = COALESCE(counterparty_id, $2),
    active_client_id = COALESCE(active_client_id, $2),
    updated_at = now()
WHERE id = $1
