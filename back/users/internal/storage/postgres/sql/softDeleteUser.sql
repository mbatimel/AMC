UPDATE users
SET status = 'deleted',
    is_active = FALSE,
    active_client_id = NULL,
    deleted_at = now(),
    updated_at = now()
WHERE id = $1
  AND deleted_at IS NULL
