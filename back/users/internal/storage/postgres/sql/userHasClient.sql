SELECT EXISTS (
    SELECT 1
    FROM user_clients uc
    JOIN users u ON u.id = uc.user_id
    WHERE uc.user_id = $1
      AND uc.client_id = $2
      AND u.deleted_at IS NULL
)
