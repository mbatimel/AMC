SELECT EXISTS (
    SELECT 1
    FROM users u
    WHERE u.deleted_at IS NULL
      AND u.is_active = TRUE
      AND (
          u.counterparty_id = $1
          OR u.active_client_id = $1
          OR EXISTS (
              SELECT 1
              FROM user_clients uc
              WHERE uc.user_id = u.id
                AND uc.client_id = $1
          )
      )
)
