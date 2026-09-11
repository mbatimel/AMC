SELECT EXISTS (
    SELECT 1
    FROM counterparties c
    WHERE regexp_replace(c.inn, '[[:space:]-]', '', 'g') = $1
      AND EXISTS (
          SELECT 1
          FROM users u
          WHERE u.deleted_at IS NULL
            AND u.is_active = TRUE
            AND (
                u.counterparty_id = c.id
                OR u.active_client_id = c.id
                OR EXISTS (
                    SELECT 1
                    FROM user_clients uc
                    WHERE uc.user_id = u.id
                      AND uc.client_id = c.id
                )
            )
      )
)
